// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/open-policy-agent/opa/compile"
	fileurl "github.com/open-policy-agent/opa/internal/file/url"
	"github.com/open-policy-agent/opa/internal/presentation"
	"github.com/open-policy-agent/opa/logging"
	"github.com/open-policy-agent/opa/metrics"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/runtime"
	"github.com/open-policy-agent/opa/util"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

// benchmarkCommandParams are a superset of evalCommandParams
// but not all eval options are exposed with flags. Only the
// ones compatible with running a benchmark.
type benchmarkCommandParams struct {
	evalCommandParams
	benchMem   bool
	count      int
	e2e        bool
	configFile string
}

const (
	benchmarkGoBenchOutput = "gobench"
)

func newBenchmarkEvalParams() benchmarkCommandParams {
	return benchmarkCommandParams{
		evalCommandParams: evalCommandParams{
			outputFormat: util.NewEnumFlag(evalPrettyOutput, []string{
				evalJSONOutput,
				evalPrettyOutput,
				benchmarkGoBenchOutput,
			}),
			target: util.NewEnumFlag(compile.TargetRego, []string{compile.TargetRego, compile.TargetWasm}),
			schema: &schemaFlags{},
		},
	}
}

func init() {
	params := newBenchmarkEvalParams()

	benchCommand := &cobra.Command{
		Use:   "bench <query>",
		Short: "Benchmark a Rego query",
		Long: `Benchmark a Rego query and print the results.

The benchmark command works very similar to 'eval' and will evaluate the query in the same fashion. The
evaluation will be repeated a number of times and performance results will be returned.

Example with bundle and input data:

	opa bench -b ./policy-bundle -i input.json 'data.authz.allow'

To enable more detailed analysis use the --metrics and --benchmem flags.

The optional "gobench" output format conforms to the Go Benchmark Data Format.
`,

		PreRunE: func(_ *cobra.Command, args []string) error {
			return validateEvalParams(&params.evalCommandParams, args)
		},
		Run: func(_ *cobra.Command, args []string) {
			exit, err := benchMain(args, params, os.Stdout, &goBenchRunner{})
			if err != nil {
				// NOTE: err should only be non-nil if a (highly unlikely)
				// presentation error occurs.
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			os.Exit(exit)
		},
	}

	// Sub-set of the standard `opa eval ..` flags
	addPartialFlag(benchCommand.Flags(), &params.partial, false)
	addUnknownsFlag(benchCommand.Flags(), &params.unknowns, []string{"input"})
	addFailFlag(benchCommand.Flags(), &params.fail, true)
	addDataFlag(benchCommand.Flags(), &params.dataPaths)
	addBundleFlag(benchCommand.Flags(), &params.bundlePaths)
	addInputFlag(benchCommand.Flags(), &params.inputPath)
	addImportFlag(benchCommand.Flags(), &params.imports)
	addPackageFlag(benchCommand.Flags(), &params.pkg)
	addQueryStdinFlag(benchCommand.Flags(), &params.stdin)
	addInputStdinFlag(benchCommand.Flags(), &params.stdinInput)
	addMetricsFlag(benchCommand.Flags(), &params.metrics, true)
	addOutputFormat(benchCommand.Flags(), params.outputFormat)
	addIgnoreFlag(benchCommand.Flags(), &params.ignore)
	addSchemaFlags(benchCommand.Flags(), params.schema)
	addTargetFlag(benchCommand.Flags(), params.target)

	// Shared benchmark flags
	addCountFlag(benchCommand.Flags(), &params.count, "benchmark")
	addBenchmemFlag(benchCommand.Flags(), &params.benchMem, true)

	// E2E flags
	addE2EFlag(benchCommand.Flags(), &params.e2e, false)
	addConfigFileFlag(benchCommand.Flags(), &params.configFile)

	RootCommand.AddCommand(benchCommand)
}

type benchRunner interface {
	run(ctx context.Context, ectx *evalContext, params benchmarkCommandParams, f func(context.Context, ...rego.EvalOption) error) (testing.BenchmarkResult, error)
}

type response struct {
	Metrics *struct {
		TimerRegoInputParseNs  *int64 `json:"timer_rego_input_parse_ns,omitempty"`
		TimerRegoPartialEvalNs *int64 `json:"timer_rego_partial_eval_ns,omitempty"`
		TimerRegoQueryEvalNs   *int64 `json:"timer_rego_query_eval_ns,omitempty"`
		TimerServerHandlerNs   *int64 `json:"timer_server_handler_ns,omitempty"`
	} `json:"metrics,omitempty"`
	Result *interface{} `json:"result,omitempty"`
}

func e2eQuery(params benchmarkCommandParams, url string, reqBody []byte, m metrics.Metrics) error {
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("expected 200 response, got %v, with body %v", resp.StatusCode, body)
	}

	var opaResp response
	if err = json.Unmarshal(body, &opaResp); err != nil {
		return err
	}

	if params.fail && !params.partial && opaResp.Result == nil {
		return fmt.Errorf("undefined result")
	}

	if params.metrics && opaResp.Metrics != nil {
		if opaResp.Metrics.TimerRegoInputParseNs != nil {
			m.Histogram("timer_rego_input_parse_ns").Update(*opaResp.Metrics.TimerRegoInputParseNs)
		}
		if opaResp.Metrics.TimerRegoQueryEvalNs != nil {
			m.Histogram("timer_rego_query_eval_ns").Update(*opaResp.Metrics.TimerRegoQueryEvalNs)
		}
		if opaResp.Metrics.TimerServerHandlerNs != nil {
			m.Histogram("timer_server_handler_ns").Update(*opaResp.Metrics.TimerServerHandlerNs)
		}
		if params.partial && opaResp.Metrics.TimerRegoPartialEvalNs != nil {
			m.Histogram("timer_rego_partial_eval_ns").Update(*opaResp.Metrics.TimerRegoPartialEvalNs)
		}
	}

	fmt.Println(string(body))

	return nil
}

func runE2E(params benchmarkCommandParams, url string, reqBody []byte) (testing.BenchmarkResult, error) {
	m := metrics.New()

	var benchErr error
	br := testing.Benchmark(func(b *testing.B) {
		m.Clear()

		if params.benchMem {
			b.ReportAllocs()
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			b.StartTimer()

			err := e2eQuery(params, url, reqBody, m)

			b.StopTimer()
			if err != nil {
				benchErr = err
				b.FailNow()
			}
		}

		for name, metric := range m.All() {
			histValues, ok := metric.(map[string]interface{})
			if !ok {
				continue
			}
			for metricName, rawValue := range histValues {
				unit := fmt.Sprintf("%s_%s", name, metricName)
				switch v := rawValue.(type) {
				case int64:
					b.ReportMetric(float64(v), unit)
				case float64:
					b.ReportMetric(v, unit)
				}
			}
		}
	})

	return br, benchErr
}

func benchE2E(ctx context.Context, args []string, params benchmarkCommandParams, w io.Writer) error {
	host := "localhost"
	port := "18181"
	addr := []string{fmt.Sprintf("%s:%s", host, port)}

	logger := logging.New()
	logger.SetLevel(logging.Error)

	dataPaths := params.dataPaths.v
	if len(params.bundlePaths.v) > 0 {
		dataPaths = append(dataPaths, params.bundlePaths.v...)
	}

	rtParams := runtime.Params{
		Addrs:                  &addr,
		Paths:                  dataPaths,
		Logger:                 logger,
		ConfigFile:             params.configFile,
		SkipBundleVerification: true,
	}

	rt, err := runtime.NewRuntime(ctx, rtParams)
	if err != nil {
		return err
	}

	initChan := rt.Manager.ServerInitializedChannel()

	cctx, cancel := context.WithCancel(ctx)
	go rt.StartServer(cctx)
	defer cancel()

	select {
	case <-initChan:
		// Server initialized, proceed with requests
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout after waiting 30 seconds for server init")
	}

	input, err := readInput(params)
	if err != nil {
		return err
	}
	query, err := readQuery(params, args)
	if err != nil {
		return err
	}

	var reqBody []byte

	// Wrap input in "input" attribute
	body := make(map[string]interface{})
	inp := make(map[string]interface{})
	if err = json.Unmarshal(input, &inp); err != nil {
		return err
	}
	body["input"] = inp

	var path string
	if params.partial {
		// TODO: test
		path = "compile"
		body["query"] = query
		if len(params.unknowns) > 0 {
			body["unknowns"] = params.unknowns
		}
	} else {
		path, err = queryToPath(query)
		if err != nil {
			return err
		}
	}

	reqBody, err = json.Marshal(body)
	if err != nil {
		return err
	}

	fmt.Println(string(reqBody))

	url := fmt.Sprintf("http://%s:%s/v1/%v", host, port, path)
	if params.metrics {
		url += "?metrics=true"
	}

	for i := 0; i < params.count; i++ {
		br, err := runE2E(params, url, reqBody)
		if err != nil {
			return renderBenchmarkError(params, err, w)
		}
		renderBenchmarkResult(params, br, w)
	}
	return nil
}

func readInput(params benchmarkCommandParams) ([]byte, error) {
	if params.stdinInput {
		return ioutil.ReadAll(os.Stdin)
	} else if params.inputPath != "" {
		path, err := fileurl.Clean(params.inputPath)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadFile(path)
	}
	return nil, nil
}

func benchMain(args []string, params benchmarkCommandParams, w io.Writer, r benchRunner) (int, error) {
	ctx := context.Background()

	if params.e2e {
		err := benchE2E(ctx, args, params, w)
		code := 0
		if err != nil {
			code = 1
		}
		return code, err
	}

	ectx, err := setupEval(args, params.evalCommandParams)
	if err != nil {
		errRender := renderBenchmarkError(params, err, w)
		return 1, errRender
	}

	var benchFunc func(context.Context, ...rego.EvalOption) error
	rg := rego.New(ectx.regoArgs...)

	if !params.partial {
		// Take the eval context and prepare anything else we possibly can before benchmarking the evaluation
		pq, err := rg.PrepareForEval(ctx)
		if err != nil {
			errRender := renderBenchmarkError(params, err, w)
			return 1, errRender
		}

		benchFunc = func(ctx context.Context, opts ...rego.EvalOption) error {
			result, err := pq.Eval(ctx, opts...)
			if err != nil {
				return err
			} else if len(result) == 0 && params.fail {
				return fmt.Errorf("undefined result")
			}
			return nil
		}
	} else {
		// As with normal evaluation, prepare as much as possible up front.
		pq, err := rg.PrepareForPartial(ctx)
		if err != nil {
			errRender := renderBenchmarkError(params, err, w)
			return 1, errRender
		}

		benchFunc = func(ctx context.Context, opts ...rego.EvalOption) error {
			result, err := pq.Partial(ctx, opts...)
			if err != nil {
				return err
			} else if len(result.Queries) == 0 && params.fail {
				return fmt.Errorf("undefined result")
			}
			return nil
		}
	}

	// Run the benchmark as many times as specified, re-use the prepared objects for each
	for i := 0; i < params.count; i++ {
		br, err := r.run(ctx, ectx, params, benchFunc)
		if err != nil {
			errRender := renderBenchmarkError(params, err, w)
			return 1, errRender
		}
		renderBenchmarkResult(params, br, w)
	}

	return 0, nil
}

type goBenchRunner struct {
}

func (r *goBenchRunner) run(ctx context.Context, ectx *evalContext, params benchmarkCommandParams, f func(context.Context, ...rego.EvalOption) error) (testing.BenchmarkResult, error) {

	var m, hist metrics.Metrics
	if params.metrics {
		m = metrics.New()
		hist = metrics.New()
	}

	ectx.evalArgs = append(ectx.evalArgs, rego.EvalMetrics(m))

	var benchErr error

	br := testing.Benchmark(func(b *testing.B) {

		// Track memory allocations, if enabled
		if params.benchMem {
			b.ReportAllocs()
		}

		// Reset the histogram for each invocation of the bench function
		hist.Clear()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {

			// Start the timer (might already be started, but that's ok)
			b.StartTimer()

			// Perform the evaluation
			err := f(ctx, ectx.evalArgs...)

			// Stop the timer while processing the results
			b.StopTimer()
			if err != nil {
				benchErr = err
				b.FailNow()
			}

			// Add metrics for that evaluation into the top level histogram
			if params.metrics {
				for name, metric := range m.All() {
					// Note: We only support int64 metrics right now, this should cover pretty
					// much all of the ones we would care about (timers and counters).
					switch v := metric.(type) {
					case int64:
						hist.Histogram(name).Update(v)
					}
				}
				m.Clear()
			}
		}

		if params.metrics {
			// For each histogram add their values to the benchmark results.
			// Note: If there are many metrics this gets super verbose.
			for histName, metric := range hist.All() {
				histValues, ok := metric.(map[string]interface{})
				if !ok {
					continue
				}
				for metricName, rawValue := range histValues {
					unit := fmt.Sprintf("%s_%s", histName, metricName)

					// Only support histogram metrics that are a float64 or int64,
					// this covers the stock implementation in metrics.Metrics
					switch v := rawValue.(type) {
					case int64:
						b.ReportMetric(float64(v), unit)
					case float64:
						b.ReportMetric(v, unit)
					}
				}
			}
		}
	})

	return br, benchErr
}

func renderBenchmarkResult(params benchmarkCommandParams, br testing.BenchmarkResult, w io.Writer) {
	switch params.outputFormat.String() {
	case evalJSONOutput:
		_ = presentation.JSON(w, br)
	case benchmarkGoBenchOutput:
		fmt.Fprintf(w, "BenchmarkOPAEval\t%s", br.String())
		if params.benchMem {
			fmt.Fprintf(w, "\t%s", br.MemString())
		}
		fmt.Fprintf(w, "\n")
	default:
		data := [][]string{
			{"samples", fmt.Sprintf("%d", br.N)},
			{"ns/op", prettyFormatFloat(float64(br.T.Nanoseconds()) / float64(br.N))},
		}
		if params.benchMem {
			data = append(data, []string{
				"B/op", fmt.Sprintf("%d", br.AllocedBytesPerOp()),
			}, []string{
				"allocs/op", fmt.Sprintf("%d", br.AllocsPerOp()),
			})
		}

		var keys []string
		for k := range br.Extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			data = append(data, []string{k, prettyFormatFloat(br.Extra[k])})
		}

		table := tablewriter.NewWriter(w)
		table.AppendBulk(data)
		table.Render()
	}
}

func renderBenchmarkError(params benchmarkCommandParams, err error, w io.Writer) error {
	o := presentation.Output{
		Errors: presentation.NewOutputErrors(err),
	}

	switch params.outputFormat.String() {
	case evalJSONOutput:
		return presentation.JSON(w, o)
	default:
		return presentation.Pretty(w, o)
	}
}

// Same format used by testing/benchmark.go to format floating point output strings
// Using this keeps the results consistent between the "pretty" and "gobench" outputs.
func prettyFormatFloat(x float64) string {
	// Print all numbers with 10 places before the decimal point
	// and small numbers with three sig figs.
	var format string
	switch y := math.Abs(x); {
	case y == 0 || y >= 99.95:
		format = "%10.0f"
	case y >= 9.995:
		format = "%12.1f"
	case y >= 0.9995:
		format = "%13.2f"
	case y >= 0.09995:
		format = "%14.3f"
	case y >= 0.009995:
		format = "%15.4f"
	case y >= 0.0009995:
		format = "%16.5f"
	default:
		format = "%17.6f"
	}
	return fmt.Sprintf(format, x)
}

func readQuery(params benchmarkCommandParams, args []string) (string, error) {
	var query string
	if params.stdin {
		bs, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		query = string(bs)
	} else {
		query = args[0]
	}
	return query, nil
}

func queryToPath(query string) (string, error) {
	if !strings.HasPrefix(query, "data.") {
		return "", fmt.Errorf("e2e query must start with 'data.'")
	}
	return strings.ReplaceAll(query, ".", "/"), nil
}
