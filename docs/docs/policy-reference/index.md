---
title: Policy Reference
sidebar_position: 3
---

## Assignment and Equality

```rego
# assign variable x to value of field foo.bar.baz in input
x := input.foo.bar.baz

# check if variable x has same value as variable y
x == y

# check if variable x is a set containing "foo" and "bar"
x == {"foo", "bar"}

# OR

{"foo", "bar"} == x
```

## Lookup

### Arrays

```rego
# lookup value at index 0
val := arr[0]

 # check if value at index 0 is "foo"
"foo" == arr[0]

# find all indices i that have value "foo"
"foo" == arr[i]

# lookup last value
val := arr[count(arr)-1]

# with keywords
some 0, val in arr   # lookup value at index 0
0, "foo" in arr      # check if value at index 0 is "foo"
some i, "foo" in arr # find all indices i that have value "foo"
```

### Objects

```rego
# lookup value for key "foo"
val := obj["foo"]

# check if value for key "foo" is "bar"
"bar" == obj["foo"]

# OR

"bar" == obj.foo

# check if key "foo" exists and is not false
obj.foo

# check if key assigned to variable k exists
k := "foo"
obj[k]

# check if path foo.bar.baz exists and is not false
obj.foo.bar.baz

# check if path foo.bar.baz, foo.bar, or foo does not exist or is false
not obj.foo.bar.baz

# with keywords
o := {"foo": false}
# check if value exists: the expression will be true
false in o
# check if value for key "foo" is false
"foo", false in o
```

### Sets

```rego
# check if "foo" belongs to the set
a_set["foo"]

# check if "foo" DOES NOT belong to the set
not a_set["foo"]

# check if the array ["a", "b", "c"] belongs to the set
a_set[["a", "b", "c"]]

# find all arrays of the form [x, "b", z] in the set
a_set[[x, "b", z]]

# with keywords
"foo" in a_set
not "foo" in a_set
some ["a", "b", "c"] in a_set
some [x, "b", z] in a_set
```

## Iteration

### Arrays

```rego
# iterate over indices i
arr[i]

# iterate over values
val := arr[_]

# iterate over index/value pairs
val := arr[i]

# with keywords
some val in arr    # iterate over values
some i, _ in arr   # iterate over indices
some i, val in arr # iterate over index/value pairs
```

### Objects

```rego
# iterate over keys
obj[key]

# iterate over values
val := obj[_]

# iterate over key/value pairs
val := obj[key]

# with keywords
some val in obj      # iterate over values
some key, _ in obj   # iterate over keys
some key, val in obj # key/value pairs
```

### Sets

```rego
# iterate over values
set[val]

# with keywords
some val in set
```

### Advanced

```rego
# nested: find key k whose bar.baz array index i is 7
foo[k].bar.baz[i] == 7

# simultaneous: find keys in objects foo and bar with same value
foo[k1] == bar[k2]

# simultaneous self: find 2 keys in object foo with same value
foo[k1] == foo[k2]; k1 != k2

# multiple conditions: k has same value in both conditions
foo[k].bar.baz[i] == 7; foo[k].qux > 3
```

## For All

```rego
# assert no values in set match predicate
count({x | set[x]; f(x)}) == 0

# assert all values in set make function f true
count({x | set[x]; f(x)}) == count(set)

# assert no values in set make function f true (using negation and helper rule)
not any_match

# assert all values in set make function f true (using negation and helper rule)
not any_not_match
```

```rego
# with keywords
any_match if {
    some x in set
    f(x)
}

any_not_match if {
    some x in set
    not f(x)
}
```

## Rules

In the examples below `...` represents one or more conditions.

### Constants

```rego
a := {1, 2, 3}
b := {4, 5, 6}
c := a | b
```

### Conditionals (Boolean)

```rego
# p is true if ...
p := true { ... }

# OR
# with keywords
p if { ... }

# OR
p { ... }
```

### Conditionals

```rego
# with keywords
default a := 1
a := 5   if { ... }
a := 100 if { ... }
```

### Incremental

```rego
# a_set will contain values of x and values of y
a_set[x] { ... }
a_set[y] { ... }

# alternatively, with keywords
a_set contains x if { ... }
a_set contains y if { ... }

# a_map will contain key->value pairs x->y and w->z
a_map[x] := y if { ... }
a_map[w] := z if { ... }
```

### Ordered (Else)

```rego
# with keywords
default a := 1
a := 5 if { ... }
else := 10 if { ... }
```

### Functions (Boolean)

```rego
# with keywords
f(x, y) if {
    ...
}

# OR

f(x, y) := true if {
    ...
}
```

### Functions (Conditionals)

```rego
# with keywords
f(x) := "A" if { x >= 90 }
f(x) := "B" if { x >= 80; x < 90 }
f(x) := "C" if { x >= 70; x < 80 }
```

### Reference Heads

```rego
# with keywords
fruit.apple.seeds = 12 if input == "apple"             # complete document (single value rule)

fruit.pineapple.colors contains x if x := "yellow"     # multi-value rule

fruit.banana.phone[x] = "bananular" if x := "cellular" # single value rule
fruit.banana.phone.cellular = "bananular" if true      # equivalent single value rule

fruit.orange.color(x) = true if x == "orange"          # function
```

For reasons of backwards-compatibility, partial sets need to use `contains` in
their rule heads, i.e.

```rego
fruit.box contains "apples" if true
```

whereas

```rego
fruit.box[x] if { x := "apples" }
```

defines a _complete document rule_ `fruit.box.apples` with value `true`.
The same is the case of rules with brackets that don't contain dots, like

```rego
box[x] if { x := "apples" } # => {"box": {"apples": true }}
box2[x] { x := "apples" } # => {"box": ["apples"]}
```

For backwards-compatibility, rules _without_ if and without _dots_ will be interpreted
as defining partial sets, like `box2`.

## Tests

```rego
# define a rule that starts with test_
test_NAME { ... }

# override input.foo value using the 'with' keyword
data.foo.bar.deny with input.foo as {"bar": [1,2,3]}}
```

## Built-in Functions

The built-in functions for the language provide basic operations to manipulate
scalar values (e.g. numbers and strings), and aggregate functions that summarize
complex types.

import BuiltinSearch from "@site/src/components/BuiltinSearch";

<BuiltinSearch entryLimit={10} />

### Comparison

<BuiltinTable category="comparison" />

### Numbers

<BuiltinTable category="numbers" />

### Aggregates

<BuiltinTable category="aggregates" />

### Arrays

<BuiltinTable category="array" />

### Sets

<BuiltinTable category="sets" />

### Objects

<BuiltinTable category="object">
- When `keys` are provided as an object only the top level keys on the object will be used, values are ignored.
  For example: `object.remove({"a": {"b": {"c": 2}}, "x": 123}, {"a": 1}) == {"x": 123}` regardless of the value
  for key `a` in the keys object, the following `keys` object gives the same result
  `object.remove({"a": {"b": {"c": 2}}, "x": 123}, {"a": {"b": {"foo": "bar"}}}) == {"x": 123}`.
- The `json` string `paths` may reference into array values by using index numbers. For example with the object
  `{"a": ["x", "y", "z"]}` the path `a/1` references `y`. Nested structures are supported as well, for example:
  `{"a": ["x", {"y": {"y1": {"y2": ["foo", "bar"]}}}, "z"]}` the path `a/1/y1/y2/0` references `"foo"`.
- The `json` string `paths` support `~0`, or `~1` characters for `~` and `/` characters in key names.
  It does not support `-` for last index of an array. For example the path `/foo~1bar~0` will reference `baz`
  in `{ "foo/bar~": "baz" }`.
- The `json` string `paths` may be an array of string path segments rather than a `/` separated string. For example
  the path `a/b/c` can be passed in as `["a", "b", "c"]`.

</BuiltinTable>

### Strings

<BuiltinTable category="strings">

:::info
When using `sprintf`, values are pre-processed and may have an unexpected type. For example,
`%T` evaluates to `string` for both `string` and `boolean` types. In such cases, use `type_name` to
accurately evaluate the underlying type.
:::

</BuiltinTable>

### Regular Expressions

<BuiltinTable category="regex"/>

### Glob Matching

<BuiltinTable category="glob">

The following table shows examples of how `glob.match` works:

| `call`                                                           | `output` | Description                                   |
| ---------------------------------------------------------------- | -------- | --------------------------------------------- |
| `output := glob.match("*.github.com", [], "api.github.com")`     | `true`   | A glob with the default `["."]` delimiter.    |
| `output := glob.match("*.github.com", [], "api.cdn.github.com")` | `false`  | A glob with the default `["."]` delimiter.    |
| `output := glob.match("*hub.com", null, "api.cdn.github.com")`   | `true`   | A glob without delimiter.                     |
| `output := glob.match("*:github:com", [":"], "api:github:com")`  | `true`   | A glob with delimiters `[":"]`.               |
| `output := glob.match("api.**.com", [], "api.github.com")`       | `true`   | A super glob.                                 |
| `output := glob.match("api.**.com", [], "api.cdn.github.com")`   | `true`   | A super glob.                                 |
| `output := glob.match("?at", [], "cat")`                         | `true`   | A glob with a single character wildcard.      |
| `output := glob.match("?at", [], "at")`                          | `false`  | A glob with a single character wildcard.      |
| `output := glob.match("[abc]at", [], "bat")`                     | `true`   | A glob with character-list matchers.          |
| `output := glob.match("[abc]at", [], "cat")`                     | `true`   | A glob with character-list matchers.          |
| `output := glob.match("[abc]at", [], "lat")`                     | `false`  | A glob with character-list matchers.          |
| `output := glob.match("[!abc]at", [], "cat")`                    | `false`  | A glob with negated character-list matchers.  |
| `output := glob.match("[!abc]at", [], "lat")`                    | `true`   | A glob with negated character-list matchers.  |
| `output := glob.match("[a-c]at", [], "cat")`                     | `true`   | A glob with character-range matchers.         |
| `output := glob.match("[a-c]at", [], "lat")`                     | `false`  | A glob with character-range matchers.         |
| `output := glob.match("[!a-c]at", [], "cat")`                    | `false`  | A glob with negated character-range matchers. |
| `output := glob.match("[!a-c]at", [], "lat")`                    | `true`   | A glob with negated character-range matchers. |
| `output := glob.match("{cat,bat,[fr]at}", [], "cat")`            | `true`   | A glob with pattern-alternatives matchers.    |
| `output := glob.match("{cat,bat,[fr]at}", [], "bat")`            | `true`   | A glob with pattern-alternatives matchers.    |
| `output := glob.match("{cat,bat,[fr]at}", [], "rat")`            | `true`   | A glob with pattern-alternatives matchers.    |
| `output := glob.match("{cat,bat,[fr]at}", [], "at")`             | `false`  | A glob with pattern-alternatives matchers.    |

</BuiltinTable>

### Bitwise Operations

<BuiltinTable category="bits" />

### Type Conversions

<BuiltinTable category="conversions" />

### Units

<BuiltinTable category="units" />

### Types

<BuiltinTable category="types" />

### Encoding

<BuiltinTable category="encoding">

The `json.marshal_with_options` builtin's `opts` parameter accepts the following properties:

| Field    | Required | Type     | Default                                                              | Description                                                                                                                                                                                                                                                                                                                                      |
| :------- | :------- | :------- | :------------------------------------------------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pretty` | No       | `bool`   | `true` if `indent` or `prefix` are declared, <br/> `false` otherwise | Enables multi-line, human-readable JSON output ("pretty-printing"). <br/>If this property is `true`, then objects will be marshaled into multi-line JSON with either user-specified or default indent/prefix options. If this property is `false`, `indent`/`prefix` will be ignored and this builtin functions identically to `json.marshal()`. |
| `indent` | No       | `string` | `"\t"` <br/> (Horizontal tab, character 0x09)                        | The string to use when indenting nested keys in the emitted JSON. One or more copies of this string will be included before child elements in every object or array.                                                                                                                                                                             |
| `prefix` | No       | `string` | `""` <br/> (empty)                                                   | The string to prefix lines with in the emitted JSON. One copy of this string will be prepended to each line.                                                                                                                                                                                                                                     |

Default values will be used if:

- `opts` is an empty object.
- `opts` does not contain the named property.

</BuiltinTable>

### Token Signing

<BuiltinTable category="tokensign">

OPA provides two builtins that implement JSON Web Signature [RFC7515](https://tools.ietf.org/html/rfc7515) functionality.

`io.jwt.encode_sign_raw()` takes three JSON Objects (strings) as parameters and returns their JWS Compact Serialization.
This builtin should be used by those that want maximum control over the signing and serialization procedure. It is
important to remember that StringOrURI values are compared as case-sensitive strings with no transformations or
canonicalizations applied. Therefore, line breaks and whitespaces are significant.

`io.jwt.encode_sign()` takes three Rego Objects as parameters and returns their JWS Compact Serialization. This builtin
should be used by those that want to use rego objects for signing during policy evaluation.

:::info
Note that with `io.jwt.encode_sign` the Rego objects are serialized to JSON with standard formatting applied
whereas the `io.jwt.encode_sign_raw` built-in will **not** affect whitespace of the strings passed in.
This will mean that the final encoded token may have different string values, but the decoded and parsed
JSON will match.
:::

The following algorithms are supported:

- `ES256`: ECDSA using P-256 and SHA-256
- `ES384`: ECDSA using P-384 and SHA-384
- `ES512`: ECDSA using P-521 and SHA-512
- `HS256`: HMAC using SHA-256
- `HS384`: HMAC using SHA-384
- `HS512`: HMAC using SHA-512
- `PS256`: RSASSA-PSS using SHA256 and MGF1-SHA256
- `PS384`: RSASSA-PSS using SHA384 and MGF1-SHA384
- `PS512`: RSASSA-PSS using SHA512 and MGF1-SHA512
- `RS256`: RSASSA-PKCS-v1.5 using SHA-256
- `RS384`: RSASSA-PKCS-v1.5 using SHA-384
- `RS512`: RSASSA-PKCS-v1.5 using SHA-512

:::info
Note that the key's provided should be base64 URL encoded (without padding) as per the specification ([RFC7517](https://tools.ietf.org/html/rfc7517)).
This differs from the plain text secrets provided with the algorithm specific verify built-ins described below.
:::

#### Token Signing Examples

##### Symmetric Key (HMAC with SHA-256)

<PlaygroundExample dir={require.context("./_examples/tokens/sign/hmac")} />

##### Symmetric Key with empty JSON payload

<PlaygroundExample dir={require.context("./_examples/tokens/sign/empty_json")} />

##### RSA Key (RSA Signature with SHA-256)

<PlaygroundExample dir={require.context("./_examples/tokens/sign/rsa")} />

##### Raw Token Signing

If you need to generate the signature for a serialized token you an use the
`io.jwt.encode_sign_raw` built-in function which accepts JSON serialized string
parameters.

<PlaygroundExample dir={require.context("./_examples/tokens/sign/sign_raw")} />

</BuiltinTable>

### Token Verification

<BuiltinTable category="tokens">

:::info
Note that the `io.jwt.verify_XX` built-in methods verify **only** the signature. They **do not** provide any validation for the JWT
payload and any claims specified. The `io.jwt.decode_verify` built-in will verify the payload and **all** standard claims.
:::

The input `string` is a JSON Web Token encoded with JWS Compact Serialization. JWE and JWS JSON Serialization are not supported. If nested signing was used, the `header`, `payload` and `signature` will represent the most deeply nested token.

For `io.jwt.decode_verify`, `constraints` is an object with the following members:

| Name     | Meaning                                                                                                                                                                                              | Required  |
| -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| `cert`   | A PEM encoded certificate, PEM encoded public key, or a JWK key (set) containing an RSA or ECDSA public key.                                                                                         | See below |
| `secret` | The secret key for HS256, HS384 and HS512 verification.                                                                                                                                              | See below |
| `alg`    | The JWA algorithm name to use. If it is absent then any algorithm that is compatible with the key is accepted.                                                                                       | Optional  |
| `iss`    | The issuer string. If it is present the only tokens with this issuer are accepted. If it is absent then any issuer is accepted.                                                                      | Optional  |
| `time`   | The time in nanoseconds to verify the token at. If this is present then the `exp` and `nbf` claims are compared against this value. If it is absent then they are compared against the current time. | Optional  |
| `aud`    | The audience that the verifier identifies with. If this is present then the `aud` claim is checked against it. **If it is absent then the `aud` claim must be absent too.**                          | Optional  |

Exactly one of `cert` and `secret` must be present. If there are any
unrecognized constraints then the token is considered invalid.

#### Token Verification Examples
The examples below use the following token:

```rego
package jwt

es256_token := "eyJ0eXAiOiAiSldUIiwgImFsZyI6ICJFUzI1NiJ9.eyJuYmYiOiAxNDQ0NDc4NDAwLCAiaXNzIjogInh4eCJ9.lArczfN-pIL8oUU-7PU83u-zfXougXBZj6drFeKFsPEoVhy9WAyiZlRshYqjTSXdaw8yw2L-ovt4zTUZb2PWMg"
```

<RunSnippet id="token.rego"/>

##### Using JWKS

This example shows a two-step process to verify the token signature and then decode it for
further checks of the payload content. This approach gives more flexibility in verifying only
the claims that the policy needs to enforce.

```rego
package jwt

jwks := `{
    "keys": [{
        "kty":"EC",
        "crv":"P-256",
        "x":"z8J91ghFy5o6f2xZ4g8LsLH7u2wEpT2ntj8loahnlsE",
        "y":"7bdeXLH61KrGWRdh7ilnbcGQACxykaPKfmBccTHIOUo"
    }]
}`
```

<RunSnippet id="jwks.rego"/>

<PlaygroundExample files="#jwks.rego #token.rego" dir={require.context("./_examples/tokens/verify/jwks")} />

The next example shows doing the token signature verification, decoding, and content checks
all in one call using `io.jwt.decode_verify`. Note that this gives less flexibility in validating
the payload content as **all** claims defined in the JWT spec are verified with the provided
constraints.

<PlaygroundExample files="#jwks.rego #token.rego" dir={require.context("./_examples/tokens/verify/jwks_single")} />

##### Using PEM encoded X.509 Certificate

The following examples will demonstrate verifying tokens using an X.509 Certificate
defined as:

```rego
package jwt

cert := `-----BEGIN CERTIFICATE-----
MIIBcDCCARagAwIBAgIJAMZmuGSIfvgzMAoGCCqGSM49BAMCMBMxETAPBgNVBAMM
CHdoYXRldmVyMB4XDTE4MDgxMDE0Mjg1NFoXDTE4MDkwOTE0Mjg1NFowEzERMA8G
A1UEAwwId2hhdGV2ZXIwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAATPwn3WCEXL
mjp/bFniDwuwsfu7bASlPae2PyWhqGeWwe23Xlyx+tSqxlkXYe4pZ23BkAAscpGj
yn5gXHExyDlKo1MwUTAdBgNVHQ4EFgQUElRjSoVgKjUqY5AXz2o74cLzzS8wHwYD
VR0jBBgwFoAUElRjSoVgKjUqY5AXz2o74cLzzS8wDwYDVR0TAQH/BAUwAwEB/zAK
BggqhkjOPQQDAgNIADBFAiEA4yQ/88ZrUX68c6kOe9G11u8NUaUzd8pLOtkKhniN
OHoCIHmNX37JOqTcTzGn2u9+c8NlnvZ0uDvsd1BmKPaUmjmm
-----END CERTIFICATE-----`
```

<RunSnippet id="cert.rego"/>

This example shows a two-step process to verify the token signature and then decode it for
further checks of the payload content. This approach gives more flexibility in verifying only
the claims that the policy needs to enforce.

<PlaygroundExample files="#cert.rego #token.rego" dir={require.context("./_examples/tokens/verify/cert")} />

The next example shows doing the same token signature verification, decoding, and content checks
but instead with a single call to `io.jwt.decode_verify`. Note that this gives less flexibility
in validating the payload content as **all** claims defined in the JWT spec are verified with the
provided constraints.

<PlaygroundExample files="#cert.rego #token.rego" dir={require.context("./_examples/tokens/verify/cert_single")} />

##### Round Trip - Sign and Verify

This example shows how to encode a token, verify, and decode it with the different options available.

Start with using the `io.jwt.encode_sign_raw` built-in:

<PlaygroundExample dir={require.context("./_examples/tokens/verify/sign_raw")} />

Now encode the and sign the same token contents but with `io.jwt.encode_sign` instead of the `raw` variant.

<PlaygroundExample dir={require.context("./_examples/tokens/verify/sign")} />

:::info
Note that the resulting encoded token is different from the first example using
`io.jwt.encode_sign_raw`. The reason is that the `io.jwt.encode_sign` function
is using canonicalized formatting for the header and payload whereas
`io.jwt.encode_sign_raw` does not change the whitespace of the strings passed
in. The decoded and parsed JSON values are still the same.
:::

</BuiltinTable>

### Time

<BuiltinTable category="time">

:::info
Multiple calls to the `time.now_ns` built-in function within a single policy
evaluation query will always return the same value.
:::

Timezones can be specified as

- an [IANA Time Zone](https://www.iana.org/time-zones) string e.g. "America/New_York"
- "UTC" or "", which are equivalent to not passing a timezone (i.e. will return as UTC)
- "Local", which will use the local timezone.

Note that OPA will use the `time/tzdata` data if none is present on the runtime filesystem (see the
[Go `time.LoadLocation()`](https://pkg.go.dev/time#LoadLocation) documentation for more information).

#### Timestamp Parsing

OPA can parse timestamps of nearly arbitrary formats, and currently accepts the same inputs as Go's `time.Parse()` utility.
As a result, either you will pass a supported constant, or you **must** describe the format of your timestamps using the Reference Timestamp that Go's `time` module expects:

    2006-01-02T15:04:05Z07:00

In other date formats, that same value is rendered as:

- January 2, 15:04:05, 2006, in time zone seven hours west of GMT
- Unix time: `1136239445`
- Unix `date` command output: `Mon Jan 2 15:04:05 MST 2006`
- RFC3339 timestamp: `2006-01-02T15:04:05Z07:00`

Examples of valid values for each timestamp field:

- Year: `"2006"` `"06"`
- Month: `"Jan"` `"January"` `"01"` `"1"`
- Day of the week: `"Mon"` `"Monday"`
- Day of the month: `"2"` `"_2"` `"02"`
- Day of the year: `"__2"` `"002"`
- Hour: `"15"` `"3"` `"03"` (PM or AM)
- Minute: `"4"` `"04"`
- Second: `"5"` `"05"`
- AM/PM mark: `"PM"`

For supported constants, formatting of nanoseconds, time zones, and other fields, see the [Go `time/format` module documentation](https://cs.opensource.google/go/go/+/master:src/time/format.go;l=9-113).

#### Timestamp Parsing Example

In OPA, we can parse a simple `YYYY-MM-DD` timestamp as follows:

<PlaygroundExample dir={require.context("./_examples/time/time_format")} />

</BuiltinTable>

### Cryptography

<BuiltinTable category="crypto"/>

### Graphs

<BuiltinTable category="graph">

A common class of recursive rules can be reduced to a graph reachability
problem, so `graph.reachable` is useful for more than just graph analysis.
This usually requires some pre- and postprocessing. The following example
shows you how to "flatten" a hierarchy of access permissions.

<PlaygroundExample dir={require.context("./_examples/graphs/reachable")} />

It may be useful to find all reachable paths from a root element. `graph.reachable_paths` can be used for this. Note that cyclical paths will terminate on the repeated node. If an element references a nonexistent element, the path will be terminated, and excludes the nonexistent node.

<PlaygroundExample dir={require.context("./_examples/graphs/reachable_paths")} />

</BuiltinTable>

### GraphQL

<BuiltinTable category="graphql">

:::info
Custom [GraphQL `@directive`](http://spec.graphql.org/October2021/#sec-Language.Directives) definitions defined by your GraphQL framework will need to be included manually as part of your GraphQL schema string in order for validation to work correctly on GraphQL queries using those directives.

Directives defined as part of the GraphQL specification (`@skip`, `@include`, `@deprecated`, and `@specifiedBy`) are supported by default, and do not need to be added to your schema manually.
:::

#### GraphQL Custom `@directive` Example

New `@directive` definitions can be defined separately from your schema, so long as you `concat` them onto the schema definition before attempting to validate a query/schema using those custom directives.
In the following example, a custom directive is defined, and then used in the schema to annotate an argument on one of the allowed query types.

```rego
package graphql_custom_directive_example

custom_directives := `
directive @customDeprecatedArgs(
  reason: String
) on ARGUMENT_DEFINITION
`

schema := `
type Query {
    foo(name: String! @customDeprecatedArgs(reason: "example reason")): String,
    bar: String!
}
`

query := `query { foo(name: "example") }`

p {
    graphql.is_valid(query,  concat("", [custom_directives, schema]))
}
```

</BuiltinTable>

### HTTP

<BuiltinTable category="http">

:::info
Similar to other built-in functions, multiple calls to the `http.send` built-in function for a given request object
within a single policy evaluation query will always return the same value.
:::

:::danger
This built-in function **must not** be used for effecting changes in
external systems as OPA does not guarantee that the statement will be executed due
to automatic performance optimizations that are applied during policy evaluation.
:::

The `request` object parameter may contain the following fields:

| Field                          | Required | Type                 | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| ------------------------------ | -------- | -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `url`                          | yes      | `string`             | HTTP URL to specify in the request (e.g., `"https://www.openpolicyagent.org"`).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `method`                       | yes      | `string`             | HTTP method to specify in request (e.g., `"GET"`, `"POST"`, `"PUT"`, etc.)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `body`                         | no       | `any`                | HTTP message body to include in request. The value will be serialized to JSON.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `raw_body`                     | no       | `string`             | HTTP message body to include in request. The value WILL NOT be serialized. Use this for non-JSON messages.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `headers`                      | no       | `object`             | HTTP headers to include in the request (e.g,. `{"X-Opa": "rules"}`).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `enable_redirect`              | no       | `boolean`            | Follow HTTP redirects. Default: `false`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `force_json_decode`            | no       | `boolean`            | Decode the HTTP response message body as JSON even if the `Content-Type` header is missing. Default: `false`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `force_yaml_decode`            | no       | `boolean`            | Decode the HTTP response message body as YAML even if the `Content-Type` header is missing. Default: `false`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `tls_use_system_certs`         | no       | `boolean`            | Use the system certificate pool. Default: `true` when `tls_ca_cert`, `tls_ca_cert_file`, `tls_ca_cert_env_variable` are unset. **Ignored on Windows** due to the system certificate pool not being accessible in the same way as it is for other platforms.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `tls_ca_cert`                  | no       | `string`             | String containing a root certificate in PEM encoded format.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `tls_ca_cert_file`             | no       | `string`             | Path to file containing a root certificate in PEM encoded format.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `tls_ca_cert_env_variable`     | no       | `string`             | Environment variable containing a root certificate in PEM encoded format.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `tls_client_cert`              | no       | `string`             | String containing a client certificate in PEM encoded format.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `tls_client_cert_file`         | no       | `string`             | Path to file containing a client certificate in PEM encoded format.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `tls_client_cert_env_variable` | no       | `string`             | Environment variable containing a client certificate in PEM encoded format.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `tls_client_key`               | no       | `string`             | String containing a key in PEM encoded format.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `tls_client_key_file`          | no       | `string`             | Path to file containing a key in PEM encoded format.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `tls_client_key_env_variable`  | no       | `string`             | Environment variable containing a client key in PEM encoded format.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `timeout`                      | no       | `string` or `number` | Timeout for the HTTP request with a default of 5 seconds (`5s`). Numbers provided are in nanoseconds. Strings must be a valid duration string where a duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". A zero timeout means no timeout.                                                                                                                                                                                                                                                                                                                                                                                         |
| `tls_insecure_skip_verify`     | no       | `bool`               | Allows for skipping TLS verification when calling a network endpoint. Not recommended for production.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `tls_server_name`              | no       | `string`             | Sets the hostname that is sent in the client Server Name Indication and that be will be used for server certificate validation. If this is not set, the value of the `Host` header (if present) will be used. If neither are set, the host name from the requested URL is used.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `cache`                        | no       | `boolean`            | Cache HTTP response across OPA queries. Default: `false`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `force_cache`                  | no       | `boolean`            | Cache HTTP response across OPA queries and override cache directives defined by the server. Default: `false`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `force_cache_duration_seconds` | no       | `number`             | If `force_cache` is set, this field specifies the duration in seconds for the freshness of a cached response.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `caching_mode`                 | no       | `string`             | Controls the format in which items are inserted into the inter-query cache. Allowed modes are `serialized` and `deserialized`. In the `serialized` mode, items will be serialized before inserting into the cache. This mode is helpful if memory conservation is preferred over higher latency during cache lookup. This is the default mode. In the `deserialized` mode, an item will be inserted in the cache without any serialization. This means when items are fetched from the cache, there won't be a need to decode them. This mode helps to make the cache lookup faster at the expense of more memory consumption. If this mode is enabled, the configured `caching.inter_query_builtin_cache.max_size_bytes` value will be ignored. This means an unlimited cache size will be assumed. |
| `cache_ignored_headers`        | no       | `list`               | List of header keys from `headers` parameter that should not considered when interacting with the cache. Default is `nil`, meaning all headers will be considered. **Important:** Note that if a cache entry exists with a subset/superset of headers that are considered in this request, it will lead to a cache miss.                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `raise_error`                  | no       | `bool`               | If `raise_error` is set, `http.send` will return an error that can halt policy evaluation when used in conjunction with the `strict-builtin-errors` option. Default: `true`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `max_retry_attempts`           | no       | `number`             | Number of times to retry a HTTP request when a network error is encountered. If provided, retries are performed with an exponential backoff delay. Default: `0`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |

If the `Host` header is included in `headers`, its value will be used as the `Host` header of the request. The `url` parameter will continue to specify the server to connect to.

When sending HTTPS requests with client certificates at least one the following combinations must be included

- `tls_client_cert` and `tls_client_key`
- `tls_client_cert_file` and `tls_client_key_file`
- `tls_client_cert_env_variable` and `tls_client_key_env_variable`

:::info
To validate TLS server certificates, the user must also provide trusted root CA certificates through the `tls_ca_cert`, `tls_ca_cert_file` and `tls_ca_cert_env_variable` fields. If the `tls_use_system_certs` field is `true`, the system certificate pool will be used as well as any additional CA certificates.
:::

The `response` object parameter will contain the following fields:

| Field         | Type     | Description                                                                                                                                                                                                                                                                                            |
| ------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `status`      | `string` | HTTP status message (e.g., `"200 OK"`).                                                                                                                                                                                                                                                                |
| `status_code` | `number` | HTTP status code (e.g., `200`). If `raise_error` is `false`, this field will be set to `0` if `http.send` encounters an error.                                                                                                                                                                         |
| `body`        | `any`    | Any value. If the HTTP response message body was not deserialized from JSON or YAML (by force or via the expected Content-Type headers `application/json`; or `application/yaml` or `application/x-yaml`), this field is set to `null`.                                                                |
| `raw_body`    | `string` | The entire raw HTTP response message body represented as a string.                                                                                                                                                                                                                                     |
| `headers`     | `object` | An object containing the response headers. The values will be an array of strings, repeated headers are grouped under the same keys with all values in the array.                                                                                                                                      |
| `error`       | `object` | If `raise_error` is `false`, this field will represent the error encountered while running `http.send`. The `error` object contains a `message` key which holds the actual error message and a `code` key which represents if the error was caused due to a network issue or during policy evaluation. |

By default, an error returned by `http.send` halts the policy evaluation when used in conjunction with the `strict-builtin-errors` option that can be set when running evaluation.
This behaviour can be altered such that instead of halting evaluation, if `http.send` encounters an error, it can return a `response` object with `status_code`
set to `0` and `error` describing the actual error. This can be activated by setting the `raise_error` field
in the `request` object to `false`. Note that if the `strict-builtin-errors` option is not specified and `raise_error`
field is `true` (which is the default), an error returned by `http.send` will generate an undefined result.

If the `cache` field in the `request` object is `true`, `http.send` will return a cached response after it checks its
freshness and validity.

`http.send` uses the `Cache-Control` and `Expires` response headers to check the freshness of the cached response.
Specifically if the [max-age](https://tools.ietf.org/html/rfc7234#section-5.2.2.8) `Cache-Control` directive is set, `http.send`
will use it to determine if the cached response is fresh or not. If `max-age` is not set, the `Expires` header will be used instead.

If the cached response is stale, `http.send` uses the `Etag` and `Last-Modified` response headers to check with the server if the
cached response is in fact still fresh. If the server responds with a `200` (`OK`) response, `http.send` will update the cache
with the new response. On a `304` (`Not Modified`) server response, `http.send` will update the headers in cached response with
their corresponding values in the `304` response.

The `force_cache` field can be used to override the cache directives defined by the server. This field is used in
conjunction with the `force_cache_duration_seconds` field. If `force_cache` is `true`, then `force_cache_duration_seconds`
**must** be specified and `http.send` will use this value to check the freshness of the cached response.

Also, if `force_cache` is `true`, it overrides the `cache` field.

`http.send` only caches responses with the following HTTP status codes: `200`, `203`, `204`, `206`, `300`, `301`,
`404`, `405`, `410`, `414`, and `501`. This is behavior is as per https://www.rfc-editor.org/rfc/rfc7231#section-6.1 and
is enforced when caching responses within a single query or across queries via the `cache` and `force_cache` request fields.

:::info
`http.send` uses the `Date` response header to calculate the current age of the response by comparing it with the current time.
This value is used to determine the freshness of the cached response. As per https://tools.ietf.org/html/rfc7231#section-7.1.1.2,
an origin server MUST NOT send a `Date` header field if it does not have a clock capable of providing a reasonable
approximation of the current instance in Coordinated Universal Time. Hence, if `http.send` encounters a scenario where current
age of the response is represented as a negative duration, the cached response will be considered as stale.
:::

The table below shows examples of calling `http.send`:

| Example                                       | Comments                                                                                                                                                                                                          |
| --------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Accessing Google using System Cert Pool       | `http.send({"method": "get", "url": "https://www.google.com", "tls_use_system_certs": true })`                                                                                                                    |
| Files containing TLS material                 | `http.send({"method": "get", "url": "https://127.0.0.1:65331", "tls_ca_cert_file": "testdata/ca.pem", "tls_client_cert_file": "testdata/client-cert.pem", "tls_client_key_file": "testdata/client-key.pem"})`     |
| Environment variables containing TLS material | `http.send({"method": "get", "url": "https://127.0.0.1:65360", "tls_ca_cert_env_variable": "CLIENT_CA_ENV", "tls_client_cert_env_variable": "CLIENT_CERT_ENV", "tls_client_key_env_variable": "CLIENT_KEY_ENV"})` |
| Unix Socket URL Format                        | `http.send({"method": "get", "url": "unix://localhost/?socket=%F2path%F2file.socket"})`                                                                                                                           |

</BuiltinTable>

### AWS

<BuiltinTable category="providers.aws">

The AWS Request Signing builtin in OPA implements the header-based auth,
single-chunk method described in the [AWS SigV4 docs](https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html).
It will default to signing the payload when present, configurable via `aws_config`, and will sign most user-provided
headers for the request, to ensure their integrity.

:::info
Note that the `authorization`, `user-agent`, and `x-amzn-trace-id` headers,
are commonly modified by proxy systems, and as such are ignored by OPA
for signing.
:::

The `request` object parameter may contain any and all of the same fields as for `http.send`.
The following fields will have effects on the output `Authorization` header signature:

| Field      | Required | Type     | Description                                                                                                                    |
| ---------- | -------- | -------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `url`      | yes      | `string` | HTTP URL to specify in the request. Used in the signature.                                                                     |
| `method`   | yes      | `string` | HTTP method to specify in request. Used in the signature.                                                                      |
| `body`     | no       | `any`    | HTTP message body. The JSON serialized version of this value will be used for the payload portion of the signature if present. |
| `raw_body` | no       | `string` | HTTP message body. This will be used for the payload portion of the signature if present.                                      |
| `headers`  | no       | `object` | HTTP headers to include in the request. These will be added to the list of headers to sign.                                    |

The `aws_config` object parameter may contain the following fields:

| Field                     | Required | Type      | Description                                                                                                                                                                                                                    |
| ------------------------- | -------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `aws_access_key`          | yes      | `string`  | AWS access key.                                                                                                                                                                                                                |
| `aws_secret_access_key`   | yes      | `string`  | AWS secret access key. Used in generating the signing key for the request.                                                                                                                                                     |
| `aws_service`             | yes      | `string`  | AWS service the request will be valid for. (e.g. `"s3"`)                                                                                                                                                                       |
| `aws_region`              | yes      | `string`  | AWS region for the request. (e.g. `"us-east-1"`)                                                                                                                                                                               |
| `aws_session_token`       | no       | `string`  | AWS security token. Used for the `x-amz-security-token` request header.                                                                                                                                                        |
| `disable_payload_signing` | no       | `boolean` | When `true` an `UNSIGNED-PAYLOAD` value will be used for calculating the `x-amz-content-sha256` header during signing, and will be returned in the response. Applicable only for `s3` and `glacier` service. Default: `false`. |

#### AWS Request Signing Examples

##### Basic Request Signing Example

The example below shows using hard-coded AWS credentials for signing the request
object for `http.send`.

:::info
For deployments, a common way to provide AWS credentials is via environment
variables, usually by using the results of `opa.runtime().env`.
:::

```rego
req := {"method": "get", "url": "https://examplebucket.s3.amazonaws.com/data"}
aws_config := {
    "aws_access_key": "MYAWSACCESSKEYGOESHERE",
    "aws_secret_access_key": "MYAWSSECRETACCESSKEYGOESHERE",
    "aws_service": "s3",
    "aws_region": "us-east-1",
}

example_verify_resource {
    resp := http.send(providers.aws.sign_req(req, aws_config, time.now_ns()))
    # process response from AWS ...
}
```

##### Unsigned Payload Request Signing Example

The [AWS S3 request signing API](https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html)
supports unsigned payload signing option. This example below shows s3 request signing with payload signing disabled.

```rego
req := {"method": "get", "url": "https://examplebucket.s3.amazonaws.com/data"}
aws_config := {
    "aws_access_key": "MYAWSACCESSKEYGOESHERE",
    "aws_secret_access_key": "MYAWSSECRETACCESSKEYGOESHERE",
    "aws_service": "s3",
    "aws_region": "us-east-1",
    "disable_payload_signing": true,
}

example_verify_resource {
    resp := http.send(providers.aws.sign_req(req, aws_config, time.now_ns()))
    # process response from AWS ...
}
```

##### Pre-Signed Request Example

The [AWS S3 request signing API](https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html)
supports pre-signing requests, so that they will only be valid at a future date.
To do this in OPA, simply adjust the time parameter:

```rego
env := opa.runtime().env
req := {"method": "get", "url": "https://examplebucket.s3.amazonaws.com/data"}
aws_config := {
    "aws_access_key": env["AWS_ACCESS_KEY"],
    "aws_secret_access_key": env["AWS_SECRET_ACCESS_KEY"],
    "aws_service": "s3",
    "aws_region": env["AWS_REGION"],
}
# Request will become valid 2 days from now.
signing_time := time.add_date(time.now_ns(), 0, 0, 2)

pre_signed_req := providers.aws.sign_req(req, aws_config, signing_time))
```

</BuiltinTable>

### Networking

<BuiltinTable category="net">

#### Notes on Name Resolution (`net.lookup_ip_addr`)

The lookup mechanism uses either the pure-Go, or the cgo-based resolver, depending on the operating system and availability of cgo.
The latter depends on flags that can be provided when building OPA as a Go library, and can be adjusted at runtime via the GODEBUG environment variable.
See [these docs on the `net` package](https://pkg.go.dev/net@go1.17.3#hdr-Name_Resolution) for details.

Note that the cgo-based resolver is often **preferable**: It will take advantage of host-based DNS caching in place.
This built-in function only caches DNS lookups within _a single_ policy evaluation.

#### Examples of `net.cidr_contains_matches`

The `output := net.cidr_contains_matches(a, b)` function allows callers to supply
strings, arrays, sets, or objects for either `a` or `b`. The `output` value in
all cases is a set of tuples (2-element arrays) that identify matches, i.e.,
elements of `b` contained by elements of `a`. The first tuple element refers to
the match in `a` and the second tuple element refers to the match in `b`.

| Input Type | Output Type   |
| ---------- | ------------- |
| `string`   | `string`      |
| `array`    | `array` index |
| `set`      | `set` element |
| `object`   | `object` key  |

If both operands are string values the function is similar to `net.cidr_contains`.

<PlaygroundExample dir={require.context("./_examples/net/cdir_contains1")} />

Either (or both) operand(s) may be an array, set, or object.

<PlaygroundExample dir={require.context("./_examples/net/cdir_contains2")} />

The array/set/object elements may be arrays. In that case, the first element must be a valid CIDR/IP.

<PlaygroundExample dir={require.context("./_examples/net/cdir_contains3")} />

If the operand is a set, the outputs are matching elements. If the operand is an object, the outputs are matching keys.

<PlaygroundExample dir={require.context("./_examples/net/cdir_contains4")} />

</BuiltinTable>

### UUID

<BuiltinTable category="uuid"/>

### Semantic Versions

<BuiltinTable category="semver">

#### Example of `semver.is_valid`

The `result := semver.is_valid(vsn)` function checks to see if a version
string is of the form: `MAJOR.MINOR.PATCH[-PRERELEASE][+METADATA]`, where
items in square braces are optional elements.

:::warning
When working with Go-style semantic versions, remember to remove the
leading `v` character, or the semver string will be marked as invalid!
:::

<PlaygroundExample dir={require.context("./_examples/semver/isvalid")} />

</BuiltinTable>

### Rego

<BuiltinTable category="rego">

#### Example

The following policy will deny the given input because:

- the `number` is greater than 5
- the `subject` does not have the `admin` role

<PlaygroundExample dir={require.context("./_examples/rego/rule_metadata")} />

#### Metadata Merge strategies

When multiple [annotations](./policy-language/#annotations) are declared along the path ancestry (chain) for a rule, how any given annotation should be selected, inherited or merged depends on the semantics of the annotation, the context of the rule, and the preferences of the developer.
OPA doesn't presume what merge strategy is appropriate; instead, this lies in the hands of the developer. The following example demonstrates how some string and list type annotations in a metadata chain can be merged into a single metadata object.

```rego
# METADATA
# title: My Example Package
# description: A set of rules illustrating how metadata annotations can be merged.
# authors:
# - John Doe <john@example.com>
# organizations:
# - Acme Corp.
package example

# METADATA
# scope: document
# description: A rule that merges metadata annotations in various ways.

# METADATA
# title: My Allow Rule
# authors:
# - Jane Doe <jane@example.com>
allow if {
	meta := merge(rego.metadata.chain())
	meta.title == "My Allow Rule" # 'title' pulled from 'rule' scope
	meta.description == "A rule that merges metadata annotations in various ways." # 'description' pulled from 'document' scope
	meta.authors == {
		{"email": "jane@example.com", "name": "Jane Doe"}, # 'authors' joined from 'package' and 'rule' scopes
		{"email": "john@example.com", "name": "John Doe"},
	}
	meta.organizations == {"Acme Corp."} # 'organizations' pulled from 'package' scope
}

allow if {
	meta := merge(rego.metadata.chain())
	meta.title == null # No 'title' present in 'rule' or 'document' scopes
	meta.description == "A rule that merges metadata annotations in various ways." # 'description' pulled from 'document' scope
	meta.authors == { # 'authors' pulled from 'package' scope
		{"email": "john@example.com", "name": "John Doe"}
	}
	meta.organizations == {"Acme Corp."} # 'organizations' pulled from 'package' scope
}

merge(chain) := meta if {
	ruleAndDoc := ["rule", "document"]
	meta := {
		"title": override_annot(chain, "title", ruleAndDoc), # looks for 'title' in 'rule' scope, then 'document' scope
		"description": override_annot(chain, "description", ruleAndDoc), # looks for 'description' in 'rule' scope, then 'document' scope
		"related_resources": override_annot(chain, "related_resources", ruleAndDoc), # looks for 'related_resources' in 'rule' scope, then 'document' scope
		"authors": merge_annot(chain, "authors"), # merges all 'authors' across all scopes
		"organizations": merge_annot(chain, "organizations"), # merges all 'organizations' across all scopes
	}
}

override_annot(chain, name, scopes) := val if {
	val := [v |
		link := chain[_]
		link.annotations.scope in scopes
		v := link.annotations[name]
	][0]
} else := null

merge_annot(chain, name) := val if {
	val := {v |
		v := chain[_].annotations[name][_]
	}
} else := null
```

</BuiltinTable>

### OPA

<BuiltinTable category="opa">

:::danger
Policies that depend on the output of `opa.runtime` may return different answers depending on how OPA was started.
If possible, prefer using an explicit `input` or `data` value instead of `opa.runtime`.
:::

### Debugging

| Built-in     | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | Details                                               |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------- |
| `print(...)` | `print` is used to output the values of variables for debugging purposes. `print` calls have no effect on the result of queries or rules. All variables passed to `print` must be assigned inside of the query or rule. If any of the `print` arguments are undefined, their values are represented as `<undefined>` in the output stream. Because policies can be invoked via different interfaces (e.g., CLI, HTTP API, etc.) the exact output format differs. See the table below for details. | <span className="tag is-warning">SDK-dependent</span> |

| API                   | Output      | Memo                                                                                                                                                                             |
| --------------------- | ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `opa eval`            | `stderr`    |                                                                                                                                                                                  |
| `opa run` (REPL)      | `stderr`    |                                                                                                                                                                                  |
| `opa test`            | `stdout`    | Specify `-v` to see output for passing tests. Output for failing tests is displayed automatically.                                                                               |
| `opa run -s` (server) | `stderr`    | Specify `--log-level=info` (default) or higher. Output is sent to the log stream. Use `--log-format=text` for pretty output.                                                     |
| Go (library)          | `io.Writer` | [https://pkg.go.dev/github.com/open-policy-agent/opa/rego#example-Rego-Print_statements](https://pkg.go.dev/github.com/open-policy-agent/opa/rego#example-Rego-Print_statements) |

</BuiltinTable>

### Tracing

<BuiltinTable category="tracing">

By default, explanations are disabled. The following table summarizes how you can enable tracing:

| API  | Parameter       | Example                                                           |
| ---- | --------------- | ----------------------------------------------------------------- |
| CLI  | `--explain`     | `opa eval --explain=notes --format=pretty 'trace("hello world")'` |
| HTTP | `explain=notes` | `curl localhost:8181/v1/data/example/allow?explain=notes&pretty`  |
| REPL | n/a             | `trace notes`                                                     |

</BuiltinTable>

## Reserved Names

The following words are reserved and cannot be used as variable names, rule
names, or dot-access style reference arguments:

```
as
contains
data
default
else
every
false
if
in
import
input
package
not
null
some
true
with
```

## Grammar

Rego’s syntax is defined by the following grammar:

```ebnf
module          = package { import } policy
package         = "package" ref
import          = "import" ref [ "as" var ]
policy          = { rule }
rule            = [ "default" ] rule-head { rule-body }
rule-head       = ( ref | var ) ( rule-head-set | rule-head-obj | rule-head-func | rule-head-comp )
rule-head-comp  = [ assign-operator term ] [ "if" ]
rule-head-obj   = "[" term "]" [ assign-operator term ] [ "if" ]
rule-head-func  = "(" rule-args ")" [ assign-operator term ] [ "if" ]
rule-head-set   = "contains" term [ "if" ] | "[" term "]"
rule-args       = term { "," term }
rule-body       = [ "else" [ assign-operator term ] [ "if" ] ] ( "{" query "}" ) | literal
query           = literal { ( ";" | ( [CR] LF ) ) literal }
literal         = ( some-decl | expr | "not" expr ) { with-modifier }
with-modifier   = "with" term "as" term
some-decl       = "some" term { "," term } { "in" expr }
expr            = term | expr-call | expr-infix | expr-every | expr-parens | unary-expr
expr-call       = var [ "." var ] "(" [ expr { "," expr } ] ")"
expr-infix      = expr infix-operator expr
expr-every      = "every" var { "," var } "in" ( term | expr-call | expr-infix ) "{" query "}"
expr-parens     = "(" expr ")"
unary-expr      = "-" expr
membership      = term [ "," term ] "in" term
term            = ref | var | scalar | array | object | set | membership | array-compr | object-compr | set-compr
array-compr     = "[" term "|" query "]"
set-compr       = "{" term "|" query "}"
object-compr    = "{" object-item "|" query "}"
infix-operator  = assign-operator | bool-operator | arith-operator | bin-operator
bool-operator   = "==" | "!=" | "<" | ">" | ">=" | "<="
arith-operator  = "+" | "-" | "*" | "/" | "%"
bin-operator    = "&" | "|"
assign-operator = ":=" | "="
ref             = ( var | array | object | set | array-compr | object-compr | set-compr | expr-call ) { ref-arg }
ref-arg         = ref-arg-dot | ref-arg-brack
ref-arg-brack   = "[" ( scalar | var | array | object | set | "_" ) "]"
ref-arg-dot     = "." var
var             = ( ALPHA | "_" ) { ALPHA | DIGIT | "_" }
scalar          = string | NUMBER | TRUE | FALSE | NULL
string          = STRING | raw-string
raw-string      = "`" { CHAR-"`" } "`"
array           = "[" term { "," term } "]"
object          = "{" object-item { "," object-item } "}"
object-item     = ( scalar | ref | var ) ":" term
set             = empty-set | non-empty-set
non-empty-set   = "{" term { "," term } "}"
empty-set       = "set(" ")"
```

The grammar defined above makes use of the following syntax. See [the Wikipedia page on EBNF](https://en.wikipedia.org/wiki/Extended_Backus–Naur_Form) for more details:

```
[]     optional (zero or one instances)
{}     repetition (zero or more instances)
|      alternation (one of the instances)
()     grouping (order of expansion)
STRING JSON string
NUMBER JSON number
TRUE   JSON true
FALSE  JSON false
NULL   JSON null
CHAR   Unicode character
ALPHA  ASCII characters A-Z and a-z
DIGIT  ASCII characters 0-9
CR     Carriage Return
LF     Line Feed
```
