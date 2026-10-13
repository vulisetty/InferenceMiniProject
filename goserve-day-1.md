# GoServe: Week 1, Day 1

## How to use this guide

You implement; your mentor explains and reviews. Work through one checkpoint at a time. Do not advance until its actual result matches the expected behavior. Expected outputs below are examples, not claims that your project has been tested.

You reported initializing a module and creating a first program in your own Git/project folder. Your mentor has not inspected that repository. Use YOUR project folder, not the separate `goserve` folder previously created by the assistant in the Scout workspace.

This guide is the only file the assistant creates for this session. No application files, dependencies, or system settings are changed by writing it.

Budget: approximately 60-90 minutes, or split this lesson into smaller sessions. No PowerShell lessons, GPU setup, Docker, or Kubernetes today.

## 1. Objective: why GoServe exists

An inference engine executes a model to produce predictions or generated tokens. A gateway handles the API-facing responsibilities around that engine.

```text
Eventually:
Client -> GoServe gateway -> model worker (vLLM) -> NVIDIA GPU
              |
              +-- validation, authentication, routing, limits, observability

Today:
Browser/client -> Go HTTP server -> health or greeting response
```

Separating these processes lets the gateway and GPU-serving runtime have different languages, dependencies, hardware requirements, and scaling policies. Gateway replicas do not each need to load model weights.

Today you will:

- Initialize or understand a Go module.
- Run a Go program.
- Implement `GET /healthz` and `GET /hello`.
- Return JSON using a struct.
- Test handlers without starting a network listener.
- Verify the real listener locally.

Afterward, explain how a request reaches a handler, how that handler writes a response, and why a health endpoint is not proof that inference works.

## 2. Development setup and prerequisites

Use VS Code on Windows with Go installed. Use VS Code's integrated terminal for commands; you do not need to learn the shell itself.

Previous environment inspection found:

| Item | Actual observation in the Scout environment |
|---|---|
| Go | `go1.27.2 windows/amd64` at `C:\Program Files\Go\bin\go.exe` |
| VS Code | Version `1.137.0` |
| Git | Version `2.55.0.vfs.0.8` |
| Docker CLI | Version `29.8.2`; engine operation/backend not verified |
| WSL | Default distribution Ubuntu; default version 2 |
| kind | Not found on PATH |
| GPU | Not checked; not needed today |

These observations do not establish what a newly opened terminal can find.

Open your project through **VS Code -> File -> Open Folder**. Choose **Terminal -> New Terminal**. Every project command in this guide runs from the folder containing `go.mod`, unless explicitly stated otherwise.

Run each command separately:

```text
go version
go env GOROOT
go env GOPATH
```

| Command | Purpose and success condition |
|---|---|
| `go version` | Confirms Go is runnable. Expect a version and platform such as `go1.27.2 windows/amd64`. |
| `go env GOROOT` | Prints the Go installation directory containing the toolchain and standard library. Previously observed: `C:\Program Files\Go`. |
| `go env GOPATH` | Prints a workspace/cache base. Previously observed: `C:\Users\svulisetty\go`. Your project does NOT have to live here. |

**If `go` is not recognized:** reopen VS Code and its terminal. A running application may have inherited PATH before Go was installed. If that still fails, diagnose PATH before reinstalling Go.

**Only if Go is genuinely absent:** use the Windows installer from [the official Go downloads page](https://go.dev/dl/) and follow [official installation instructions](https://go.dev/doc/install). Choose the installer matching your Windows architecture, then reopen your terminal and repeat the checks. Do not upgrade an existing installation just to follow this lesson.

**Checkpoint:** record the three actual outputs. No tool installations are necessary if these work.

## 3. Module: the project's identity

A **package** groups Go source code. A **module** groups packages under one identity and dependency/version configuration.

`go.mod` is roughly comparable to the identity and dependency portion of a `.csproj`. It does not enumerate your source files.

If your project already has `go.mod`, inspect it and skip initialization.

Otherwise, from your intended project root:

```text
go mod init goserve
```

Prerequisites: Go works, and this is the folder you want as the module root.

Expected output:

```text
go: creating new go.mod: module goserve
```

The generated file will resemble:

```go
module goserve

go 1.27.2
```

The Go version depends on your installed toolchain. The `go` directive declares the minimum required Go version. This lesson needs Go 1.22 or newer for method-aware HTTP routes.

`goserve` is fine for this local learning project. When publishing, a repository-aligned module path such as `github.com/YOUR-USERNAME/goserve` is conventional; use your actual repository path then, not a guessed identity now.

Standard-library packages such as `fmt` and `net/http` do not need dependency entries.

**Common error:** `go.mod already exists` means initialization already happened. Do not delete the file or create another nested module.

**Checkpoint:** `go.mod` exists in the project root. Explain the difference between Git tracking your changes and Go tracking your module.

## 4. First program: where execution starts

If you already ran this successfully, read the explanation and skip rewriting it.

Create `main.go` beside `go.mod`:

```go
package main

import "fmt"

func main() {
	fmt.Println("GoServe is starting")
}
```

### Read the code

- `package main`: this file belongs to the executable program's package. A package is not exactly a C# namespace.
- `import "fmt"`: makes the standard formatting package available in this file. Imports are per file.
- `func`: declares a function.
- `main()`: the executable's entry point. Unlike C#'s common `Main` signatures, Go's `main` takes no parameters and returns no value.
- `{ ... }`: the function body.
- `fmt.Println`: like `Console.WriteLine`; prints to standard output with a newline.
- Go normally inserts semicolons for you. Keep the opening function brace on the same line.

Save the file, then run:

```text
go run .
```

This compiles and runs the package in the current directory. The dot means "the current directory." The command does not leave a named application executable in your project.

Expected output:

```text
GoServe is starting
```

Then the program exits because `main` finishes.

**Common errors:** "no Go files" usually means the file is unsaved or the terminal is in the wrong directory. An unused import is a compiler error; remove imports you are not using.

**Your exercise:** add `fmt.Println("Ready")`. Predict the order, then run it.

**Checkpoint:** explain why this program exits instead of staying alive.

## 5. First HTTP endpoint

### Concepts before code

**HTTP server:** a program listening for HTTP requests on a network address.

**Port:** a numbered endpoint on a machine. We will use port 8080.

**Loopback:** `127.0.0.1` refers to your machine. Binding here avoids exposing this learning server to other machines.

**Router:** chooses a handler based on a request's method and path.

**Handler:** a function that reads a request and writes a response. Similar to the delegate you register with ASP.NET Core `MapGet`.

**Response:** a status code, headers, and an optional body. Status `200` means successful handling; it does not guarantee the whole platform is healthy.

Go's standard `net/http` package provides these basics without a third-party framework.

Replace `main.go` with:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := fmt.Fprintln(w, "ok"); err != nil {
		log.Printf("write health response: %v", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)

	log.Println("GoServe listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", mux))
}
```

### Read the important blocks

`healthHandler` receives two parameters:

- `w http.ResponseWriter`: an interface supplied by Go's HTTP server. It gives us operations for writing a response; think of a contract somewhat like a C# interface. We do not implement one today.
- `r *http.Request`: a pointer to the incoming request. For now, understand that we receive access to an existing request rather than copying its struct. Go pointers are not C# `ref` parameters; we will study them separately.

Unused function parameters are allowed, so `r` can remain unused here.

`Header().Set` declares the response body's type. Headers must be set before writing the body.

`fmt.Fprintln(w, "ok")` sends text to the response writer rather than the terminal. Its two results are a byte count and an error. `_` discards the count.

`if _, err := ...; err != nil` declares results and checks the error in one statement. Go commonly returns errors as values rather than throwing exceptions. `nil` means no error here. The `err` variable is scoped to this `if`.

Without an explicit `WriteHeader`, the first body write implicitly sends `200 OK`. If writing fails, the response may already be committed; log the error rather than trying to send a replacement response.

`http.NewServeMux()` creates a router. `:=` declares a local variable and infers its type.

`HandleFunc("GET /healthz", healthHandler)` registers the function without calling it now. Parentheses would call it; passing its name gives the router the function to invoke later.

`ListenAndServe` listens and serves requests until it encounters an error. It does not immediately finish like `Println`.

`log.Fatal` logs an error and exits. This is a small-demo startup pattern, not graceful shutdown. The preceding "listening" message is an announcement, not proof that binding succeeded; an HTTP request is the real check.

We use an explicit router rather than global registration to keep ownership local and make testing simpler.

### Run it

From your project root in VS Code's terminal, run individually:

```text
go fmt ./...
go vet ./...
go run .
```

| Command | Why and success condition |
|---|---|
| `go fmt ./...` | Applies standard Go formatting to packages under the module. May print filenames it changed; otherwise no output. Uses Go's standard formatter, `gofmt`. |
| `go vet ./...` | Checks for common suspicious code, such as incorrect formatting arguments. Normally silent on success. Not a behavioral test. |
| `go run .` | Starts the server. Terminal stays occupied; no startup error should appear. |

Open `http://localhost:8080/healthz` in your browser. Expect `ok`.

If it cannot connect, inspect the server terminal. If port 8080 is already occupied, stop your previous server using Ctrl+C. Do not terminate an unknown process. If necessary, choose another unused port and update both code and URL.

**Checkpoint:** the real endpoint responds. Why does the terminal remain occupied?

Stop the server with Ctrl+C before editing the next step. `go run` does not automatically reload changed files.

## 6. Add `/hello` and a JSON response

### Concepts before code

**JSON:** a text format clients can parse into structured data.

**Struct:** a Go type grouping named fields. Unlike a C# class, a struct is a value type and does not provide class inheritance.

**Exported field:** a field whose name starts with an uppercase letter. Go's JSON encoder needs exported fields to read them.

**Struct tag:** metadata after a field declaration. The JSON encoder uses it to choose the field name in JSON.

For our greeting, a struct defines the response shape explicitly. We choose it over a generic map because the expected fields are clear and compiler-checked.

Your final `main.go` for today:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type helloResponse struct {
	Message string `json:"message"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := fmt.Fprintln(w, "ok"); err != nil {
		log.Printf("write health response: %v", err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := helloResponse{
		Message: "Hello from GoServe",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("write hello response: %v", err)
	}
}

func newRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /hello", helloHandler)
	return mux
}

func main() {
	log.Println("GoServe listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", newRouter()))
}
```

### What changed

- `encoding/json` is Go's built-in JSON package.
- `type helloResponse struct` defines our response type. The lowercase type name keeps it unexported; the uppercase `Message` field is exported.
- `Message string` means the field holds a string.
- The backtick tag `json:"message"` makes the JSON key lowercase.
- `helloResponse{Message: ...}` constructs a struct value.
- `json.NewEncoder(w).Encode(response)` serializes the value directly into the response body, with a trailing newline.
- `newRouter()` constructs our routes separately from starting the listener. Its return type, `*http.ServeMux`, is a pointer to the router.
- Each HTTP request gets its own local response value. We have no shared mutable application state.

Incorrect example:

```go
type helloResponse struct {
	message string `json:"message"`
}
```

Even with a tag, lowercase `message` is unexported. Encoding this struct produces `{}`, not the intended field.

### Request lifecycle

```text
GET /hello
    |
    v
TCP connection to 127.0.0.1:8080
    |
    v
net/http parses the HTTP request
    |
    v
ServeMux matches GET + /hello
    |
    v
helloHandler sets Content-Type and encodes the struct
    |
    v
Client receives 200 + JSON body
```

ASP.NET Core analogy: `HandleFunc` plays a role like endpoint registration, `r` resembles request information from `HttpContext`, and `w` exposes response-writing operations. Go does not automatically serialize a returned object here; our handler explicitly writes the response.

The HTTP server manages concurrent handling internally. We do not need to create goroutines for these endpoints. This does NOT mean every future shared variable will be safe without synchronization.

## 7. Verify the real endpoints

Run formatting and static checks again, then start the server:

```text
go fmt ./...
go vet ./...
go run .
```

Use these browser URLs:

| URL | Expected behavior |
|---|---|
| `http://localhost:8080/healthz` | `ok` |
| `http://localhost:8080/hello` | `{"message":"Hello from GoServe"}` |
| `http://localhost:8080/missing` | `404 page not found` |

For the original assignment's PowerShell verification, open a SECOND VS Code terminal in the same project. Copy these commands; no shell lesson is needed:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/healthz" -Method Get
Invoke-RestMethod -Uri "http://localhost:8080/hello" -Method Get
```

Prerequisite: the server is still running in the first terminal.

The first prints `ok`. The second parses JSON and shows a `message` property with `Hello from GoServe`; its display may look like a table rather than raw JSON.

To inspect status and headers:

```powershell
Invoke-WebRequest -Uri "http://localhost:8080/hello" -Method Get |
    Select-Object StatusCode, Headers, Content
```

Expect status `200`, a JSON content type, and the greeting body. These commands contact only your local server.

Connection errors mean the listener is unavailable or the port/URL differs. An HTTP error means a server responded but rejected or could not handle that request.

**Checkpoint:** record actual responses. A browser display alone does not verify response headers; the inspection command or browser developer tools can.

## 8. Add automated tests alongside the feature

### Concepts before code

A **unit/handler test** calls our router in memory without opening a port.

`httptest.NewRequest` builds an incoming request for a test. `httptest.NewRecorder` captures the status, headers, and body our handler writes.

This proves routing and response behavior. It does NOT prove that a network port binds, a container starts, or a GPU worker is healthy.

Create `main_test.go` beside `main.go`:

```go
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	newRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/plain; charset=utf-8", got)
	}
	if got := recorder.Body.String(); got != "ok\n" {
		t.Errorf("body = %q, want %q", got, "ok\n")
	}
}

func TestHello(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/hello", nil)
	recorder := httptest.NewRecorder()

	newRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var response helloResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Message != "Hello from GoServe" {
		t.Errorf("message = %q, want %q", response.Message, "Hello from GoServe")
	}
}

func TestHelloRejectsPost(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/hello", nil)
	recorder := httptest.NewRecorder()

	newRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}

func TestUnknownRoute(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)
	recorder := httptest.NewRecorder()

	newRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
```

### Understand the test code

- Go discovers files ending in `_test.go`.
- Test functions start with `Test` and take `*testing.T`.
- `nil` in `NewRequest` means no request body.
- `ServeHTTP` invokes the router directly.
- `t.Fatalf` reports failure and stops this test; `t.Errorf` reports failure but lets it continue.
- `%d`, `%q`, and `%v` format an integer, quoted text, and a general value.
- `var response helloResponse` declares a variable with its zero value.
- `&response` gives the JSON decoder a pointer so it can populate that variable. Without a pointer, it cannot fill the caller's struct.

`GET` route patterns also match `HEAD` under Go's router rules; they are not an absolute "GET only" rule. A real HTTP server suppresses the response body for HEAD. POST should return `405 Method Not Allowed`.

From the project root, run individually:

```text
go fmt ./...
go vet ./...
go test ./...
go test -v ./...
```

Prerequisites: both source files are saved and compile. Tests can run without a running server.

Expected success signals:

- Formatting and vet finish without errors.
- `go test ./...` reports `ok` for your package.
- Verbose tests show all four test names passing.

`./...` means the current package and packages below it, not every Go project on the machine.

If you see `[no test files]`, check the filename and folder. If a test fails, compare its reported actual and expected values; do not change the expectation merely to make it pass.

### Race detection: know it, don't install tools blindly

```text
go test -race ./...
```

The race detector looks for unsynchronized concurrent memory access on exercised code paths. It is not a proof that all possible races are absent.

On Windows, it requires a supported C compiler toolchain and CGO enabled, even though our application itself uses only Go's standard library. If it fails because those prerequisites are unavailable, record **"race check blocked by local toolchain prerequisites"**. Do not install a compiler or change system configuration today solely to silence that error.

Our current tests do not generate concurrent load, so a passing race check would provide limited concurrency evidence.

## 9. Participate: small independent exercises

Try before reading the hints.

1. Change `/hello` to return `Hello, inference engineer`. Update the matching assertion, format, and rerun the tests. Restart the server and check the browser.
2. Temporarily make `Message` lowercase in the struct and update its uses so the program compiles. Predict the JSON, then run the test. Restore the exported field afterward.
3. Add a test showing that `POST /healthz` returns 405, using the existing POST test as a pattern.

Hints:

- A running `go run` process does not reload edits.
- JSON needs an exported field, not just a JSON tag.
- For exercise 3, create a request, invoke the router with a recorder, and compare the status.

Understanding questions:

- Why does `main` stay running in the HTTP version?
- What is the difference between `fmt.Println` and `fmt.Fprintln(w, ...)`?
- Why set headers before writing a response body?
- Why does `go test` not require port 8080 to be free?
- What does `/healthz` prove, and what does it not prove?

## 10. Design boundaries: this is not production-ready

Today we deliberately keep a small standard-library implementation.

Missing production capabilities include configurable binding, HTTP timeouts, graceful shutdown, request/body limits, authentication, structured errors/logs, real worker health, routing, inference, and usage metering.

Port 8080 and the loopback address are fixed for this isolated exercise. Later, deployment configuration will supply them. Loopback binding will also need reconsideration inside containers.

`/healthz` checks that this process responds, not readiness to serve a model. We will distinguish liveness and readiness when a worker dependency exists.

No claims about latency, throughput, Kubernetes readiness, GPU compatibility, or model quality can be made from today's endpoints.

## 11. Finish the session: evidence and learning log

Final intended project files:

```text
YOUR-PROJECT\
    go.mod
    main.go
    main_test.go
```

No `go.sum` is required for this standard-library-only project.

Stop the server with Ctrl+C when done.

Fill in this log with your actual results, either in a copy of this guide or your own notes. Leave unknowns explicitly unknown.

| Item | Your record |
|---|---|
| Repository/project path | |
| Go version | |
| Module name | |
| `/healthz` actual response | |
| `/hello` actual response | |
| Formatting result | |
| Static analysis result | |
| Test result and number of tests | |
| Race check result or prerequisite blocker | |
| Exercises completed | |
| Unresolved errors/questions | |
| Git commit identifier, if committed | |

Session recap:

1. **Built:** a local Go HTTP server with health and JSON greeting endpoints, plus handler tests.
2. **Learned:** modules, packages, imports, functions, structs, exported fields, JSON tags, basic error values, routing, and response writing.
3. **Remember:** `go run .`, `go fmt ./...`, `go vet ./...`, and `go test ./...`.
4. **Verification:** fill in actual results above; completion is not assumed.
5. **Unresolved issues:** record errors rather than skipping them.
6. **Independent exercise:** change the greeting and add the health POST test.
7. **Next session:** review your code and test evidence; then introduce explicit server configuration and a reproducible build. Do not start Docker or Kubernetes from this guide.
8. **Suggested commit message:** `feat: add GoServe health and hello endpoints with handler tests`.

### Optional Git checkpoint

Only after the behavior works, from your project folder inside your existing Git repository:

```text
git status --short
git diff -- main.go main_test.go go.mod
```

These inspect changes. `git diff` does not show contents of untracked files; review those in VS Code. Confirm these are your intended project files and do not include credentials or unrelated work.

Then stage only these files:

```text
git add go.mod main.go main_test.go
git diff --cached
```

Review the staged diff. If unrelated files were already staged, do not blindly commit them; use VS Code's Source Control UI to select only the intended changes.

When the staged contents are exactly what you intend:

```text
git commit -m "feat: add GoServe health and hello endpoints with handler tests"
```

Success prints a commit summary. If Git reports missing author identity, configure it according to your existing personal/work repository policy; do not invent an identity.

This creates a local commit only. It does not push or publish your project.
