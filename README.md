# Blueprint

## __Installation Instructions__

To install Blueprint, you only need a brief set of dependencies that are highlighted below.

__Note__: Python and Pip refer to Python3 and Pip3 respectively. Blueprint expects that the command ```python``` points to a python3 installation.

### __Dependencies__

+ Go version >= 1.18
+ Python version >=3.8
+ Thrift: [install](https://thrift.apache.org/docs/install/debian.html), [download](https://thrift.apache.org/download)
+ Grpc for go: [instructions](https://grpc.io/docs/languages/go/quickstart/)
+ Kompose: [install](https://kompose.io/)

To download the go dependencies, please execute the following commands

```
> go mod tidy
> go mod download
```

In addition, you need some python libraries which can be downloaded using the following commands

```
> pip install astunparse
```

To deploy applications, you will need to install ```docker``` and ```docker-compose``` on all the machines you intend to deploy the services on.
Note that for newer docker versions, ```compose``` is a built-in command does not require additional instructions.
If using newer versions of docker, replace occurrences of ```docker-compose``` in the following with ```docker compose```.

## __Running Blueprint__

Generating an application with Blueprint from a specification requires a config file as an input.

The command for running Blueprint is as follows:

```
> go build ./cmd/blueprint
> ./blueprint -config=<path/to/config.json> [-verbose]
```

For running the generated applications on your local machine, please install [docker-compose](https://docs.docker.com/compose/install/)

### __Config File__

The config file has the following required options:

+ 'app_name' : Name of the application
+ 'src_dir' : Path to the directory containing the application specification
+ 'output_dir' : Path to the output directory where the generated system will be placed.
+ 'wiring_file' : Path to the wiring file for the application
+ 'target' : The language in which the system should be generated. Currently, only golang [go] is supported.
+ 'deploy_address' : The global IP address to be used for the generated services.

The config file also has a way for overriding the global address options for service instances as well as the ports for each services. These are not required. Here is an example of how to override ports and addresses. Note that the name of the service MUST match the service variable in the wiring file.

```
'addresses' : [
    {
        'name' : 'serviceA',
        'address' : 'serviceA',
        'port': 9000
    },
    {
        'name' : 'serviceB',
        'port': 9001
    }
]
```

#### __Physical Machines + Clusters__

Blueprint provides choices that allow generated systems to be easily deployed at specific machines in a given cluster. For this, the config file must provide an inventory list of all physical machines available in the cluster. To do so, use the 'inventory' option. In the following example, we have added an inventory consisting of 2 machines with the first machine serving as the machine where the system will be built. We have also modified the address information of 'serviceA' to have a hostname which matches one of the machines in the inventory.

```
'inventory' : [
    {"hostname": "pinky05", "is_build_node": true},
    {"hostname": "pink04"}
],
'addresses' : [
    {
        'name' : 'serviceA',
        'address' : 'serviceA',
        'port': 9000,
        'hostname': "pinky05"
    },
    {
        'name' : 'serviceB',
        'port': 9001
        'hostname': "pinky04"
    }
]
```

A sample config file can be found at [examples/Leaf/input/config_go.json](examples/Leaf/input/config_go.json)

## __Off-The-Shelf Use__

Blueprint ships with multiple applications that are ready to use and/or modify. Most of the applications are reproductions of applications and systems from existing Microservice Benchmark Suites such as DeathStarBench.

### __Starter Applications__

Note: The docker-compose commands might require 'sudo' privileges depending on how docker-compose is installed on your machine.

#### __Leaf Application__

Leaf application is the simplest application that Blueprint ships with.
Comprising of only 3 services, this application's primary goal is to show users
how to write an application consisting of services that call each other.

The Leaf application is available at [examples/Leaf](examples/Leaf).

To build and launch the Leaf app, execute the following commands:

```
> go run main.go -config=examples/Leaf/input/config_go_grpc.json
> cd examples/Leaf/output_go
> docker-compose build
> docker-compose up -d
```

The default leaf application consists of 2 services communicating over RPC with a frontdoor service acting as a web gateway. 

To test that the application is running correctly, execute the following command:

```
> curl localhost:9502/leaf?a=5
```

For the internal services, the user must connect to them via a RPC client library. For the GRPC deployment, we recommend using [ghz](https://github.com/bojand/ghz) as the workload client.

#### __Toy Application__

A simple application intended as an example application for onboarding new users of Blueprint.
This application is a bit more complex than the Leaf application as it introduces the notion of components.
Toy application consists of 3 services which use a database, a cache, and a queue.

The Toy application is available at [examples/Toy/](examples/Toy).

To build and run the toy example, execute the following commands:

```
> go run main.go -config=examples/Toy/input/config_go.json
> cd examples/Toy/output_go
> docker-compose build
> docker-compose up -d
```

By default, the frontdoor web-server will bind to [http://localhost:9000](http://localhost:9000) and will export the apis "/foo?a=<int>", "/getb", "setb?b=<int>", and "/contactc". You can send requests to the server using curl:

```
> curl localhost:9000/foo?a=5
```

The toy example ships with a jaeger tracer with traces available for viewing at [http://localhost:16686](http://localhost:16686).

### __Benchmark Applications__

Blueprint provides 5 out of the box benchmark applications. For details about building and running individual applications, please visit their individual repositories:

+ [DeathStarBench-SocialNetwork](https://gitlab.mpi-sws.org/cld/blueprint/systems/blueprint-dsb-socialnetwork)
+ [DeathStarBench-HotelReservation](https://gitlab.mpi-sws.org/cld/blueprint/systems/blueprint-dsb-hotelreservation)
+ [DeathStarBench-MediaMicroservices](https://gitlab.mpi-sws.org/cld/blueprint/systems/blueprint-dsb-media)
+ [TrainTicket](https://gitlab.mpi-sws.org/cld/blueprint/systems/blueprint-trainticket)
+ [SockShop](https://gitlab.mpi-sws.org/cld/blueprint/systems/blueprint-sockshop)

### __Workload Generator__

Blueprint provides an out of the box HTTP Workload Generator that can generate a configurable WorkLoad for any of the aforementioned applications.

#### __Usage__

Once you have deployed the application, use the following command to run the workload generator

```bash
> cd cmd/wrk_http
> go build
> ./wrk_http -config=path/to/config.json -tput=10000 -duration=1m -outfile=stats.csv
```

#### __Command Line Options__

+ ```config```: Path to config file describing the list of APIs and their respective request parameter generation functions
+ ```tput```: Desired throughput (in reqs/s)
+ ```duration```: Duration for which the workload should execute. Must be a valid Go time string
+ ```outfile```: Path to file where the measured statistics will be written.

#### __Config Options__

The config file consists of the following 2 options

+ ```url```: Full url where the front-end web-server is hosted
+ ```apis```: List of all APIs that need to be executed

Each API must have the following fields

+ ```name```: The name of the endpoint
+ ```arg_gen_func_name```: The name of the function ued to generate the request parameters
+ ```proportion```: The percentage of all requests that should be for this API. This value should be an integer between 0 and 100.
+ ```type```: Type of the request. One of ```POST``` or ```GET```.

Note that the sum of all proportions for APIs should be exactly 100.
An example config can be found at [cmd/wrk_http/hotel_Blueprint.json](cmd/wrk_http/hotel_Blueprint.json).

#### __Request Parameter Generation__

Generating the workload for an application requires writing application-specific request-generators which specify how to generate the arguments
for the various APIs provided by the application.

All bindings must be added to the [workload](workload) folder.

Here is an example of the request-generator for the ```ReadUserTimeline``` API provided by the Social Network application.

```go
import (
    "net/url"
    "rand"
)

// Workload binding for ReadUserTimeline API
func ReadUserTimeline(is_original bool) url.Values {
	user_id := rand.Intn(962)
	start := rand.Intn(100)
	stop := start + 10
	data := url.Values{}
	data.Add("user_id", prepareArg(user_id, is_original))
	data.Add("start", prepareArg(start, is_original))
	data.Add("stop", prepareArg(stop, is_original)) 
	return data
}
```

The function must be registered as a valid API request generator with a unique ID that can be used by workload configuration files.
To register the function, modify the ```NewWorkloadRegistry``` in the [workload/registry.go] file to register the function as a valid request generator that can be used
by the workload generator.

### __Existing Features__

Currently, Blueprint supports the following:

+ __Cache__ - Memcached, Redis
+ __Tracer__ - Jaeger, Zipkin
+ __NoSQLDatabase__ - MongoDB
+ __RelationalDatabase__ - MySQL
+ __Queue__ - RabbitMQ
+ __Rpc Frameworks__ - Thrift, grpc
+ __Xtrace__

## __Modifying Applications__

There are multiple ways one could modify existing application. We highlight some of the more common ways of doing so below.

### __Adding a new Service__
 
Let's say you wanted to add a new service called FooBarService that only has 1 remote function called foobar. This function takes as input 1 argument, calls a function foo in serviceA and a function bar in serviceB, and then returns an integer as the result. Then that service will be specified in the specification as 2 different parts - (i) an interface with all the method signatures; (ii) an implementation of the interface.

The interface of the service represents the list of methods along with their signatures that a client or a server needs to implement to satisfy the criteria.The implementation of the service represents the core functionality for each API exported by the service. Blueprint uses the interface of a service to generate the clients that other services use to connect to this service.

```go
import "context"

// FooBarService interface
type FooBarService interface {
    FooBar(ctx context.Context, arg1 int) (int, error)
}
```

The code above shows the `interface` part of the service specification. Each API to be exported by the interface must satisfy the following 2 criteria:

+ The first parameter to the function must be of type `context.Context` provided by the standard `context` package
+ The last return value of the function should be of type `error`.

```go
// FooBarService Core Code
type FooBarServiceImpl struct {
    serviceA ServiceA
    serviceB ServiceB
}

func NewFooBarServiceImpl(serviceA ServiceA, serviceB ServiceB) *FooBarServiceImpl {
    return &FooBarServiceImpl{serviceA: serviceA, serviceB: serviceB}
}

func (f * FooBarServiceImpl) FooBar(ctx context.Context, arg1 int) (int, error) {
    val, err := f.serviceA.Foo(ctx, arg1)
    if err != nil {
        return -1, err
    }
    res, err := f.serviceB.Bar(val)
    if err != nil {
        return -1, err
    }
    return res, nil
}
```

The code above shows the `implementation` part of the service specification. In addition to method implementations, each service implementation must provide a `constructor` method that takes as arguments the other services, the implementation depends on. Here, `FooBarServiceImpl` depends on `ServiceA` and `ServiceB` so its constructor takes as arguments `serviceA` of type `ServiceA` and `serviceB` of type `ServiceB`.

### __Adding a remote data type__

For the specification language, a remote data type is any custom class that will be sent over the network as a single entity. To define such a data type, the struct must implement the `Remote` interface defined in [stdlib/types/remote.go](remote.go). The fields of the struct definition for the data type define the list of members for the data type. A remote data type can only be made up of base types or other remote data types. Non-remote user-defined data types are not allowed to be members of remote data types.

Here is an example from the DeathStarBench specification for the `User` data type.

```go
type User struct {
	UserID int64
	FirstName string
	LastName string
	Username string
	PwdHashed string
	Salt string
}

func (u User) remote() {}
```

### __Modifying the Wiring File and Config Options__

Each application has a corresponding wiring file and config file. Changing various options in those files can change the deployment and concrete implementations of various components as well as the server and client implementations for each service. These modifications are independent of the application specification.

Here is an example of some common modifications a user might want to do in the wiring file

#### __Changing the RPC framework__

```diff
-rpc_server : Modifier = RPCServer(framework="grpc", timeout="500ms")
+rpc_server : Modifier = RPCServer(framework="aiothrift", timeout="500ms")


fooService : Service = FooServiceImpl().WithServer(rpc_server)
```

#### __Adding Tracing__

```diff
rpc_server : Modifier = RPCServer(framework="grpc", timeout="500ms")

+jaegerTracer : Tracer = JaegerTracer().WithServer(normal_deployer)
+jaegerTracerModifier: Callable[str, Modifier] = lambda x : TracerModifier(tracer=jaegerTracer, service_name=x)
-server_modifiers : Callable[str, List[Modifier]] = lambda x : [rpc_server]
+server_modifiers : Callable[str, List[Modifier]] = lambda x : [rpc_server, jaegerTracerModifier(x)]

fooService : Service = FooServiceImpl().WithServer(server_modifiers("FooService"))
```

## __Adding a new Application__

When adding a new application, we recommend adding the application in the examples folder by creating a new folder for the application. Then we recommend the following folder structure:

```
    + examples
        + new_app
            + input
                + input_go
                    + services
                        - ServiceA.go
                        - ServiceB.go
                    - go.mod
                - config.json
            + output_go
            + wiring
                - instances.py
```

Here, the new_app folder is the new created folder for the application which contains 3 folders - (i) input folder containing the configuration file as well as the service specifications; (ii) output folder where the generated system will be placed; (iii) wiring folder where the wiring file will be placed

## __Extending Blueprint__

Blueprint is designed to be extensible with new fetaures.

We envision that users can add all sorts of features!
How one decides to implement a new feature is upto the person adding the feature. To support such extensibility, we are providing the steps for extending various parts of Blueprint. We believe that adding a new feature will consist of one or more of the type of changes highlighted below. Note that this list is not meant to be exhaustive but to cover the majority of the features we expect users might add.

### __Adding new Components and Choices__

Components represent reusable parts of the system that user-defined services depend upon. These components provide a generic interface that services can program against. These component interfaces are instantiated by one of the concrete choices for that component.

#### __Adding a new Component__

Each Component exports an interface that service specifications program against. To add a new component, follow the following steps:

+ In [stdlib/components](stdlib/components), create a new go file for the component.
+ In the file, add an interface for the component!
+ In [wiring/registry.py](wiring/registry.py), add the new component to the list of Components so that the wiring parser can correctly identify components.

Here is an example of the `Cache` interface

```go
type Cache interface {
	Put(key string, value interface{}) error
	// val is the pointer to which the value will be stored
	Get(key string, val interface{}) error
	Mset(keys []string, values []interface{}) error
	// values is the array of pointers to which the value will be stored 
	Mget(keys []string, values []interface{}) error
	Delete(key string) error
	Incr(key string) (int64, error)
}
```

#### __Adding a new Choice__

Adding a new choice for a component requires doing 2 things: (i) Adding an implementation of the `Component` interface; (ii) Extending the Blueprint IR to have knowledge of the new choices.

To add an implementation of the `Component` interface, follow the following steps:

+ In [stdlib/choices](stdlib/choices), find the component folder for which you want to add a new choice.
+ If the folder exists, skip to the next step. If the folder does not exist, create an empty folder for the component.
+ Inside the folder, create a new go file and add the implementation for the choice. See [stdlib/choices/cache/redis.go](stdlib/choices/cache/redis.go) for the `Redis` implementation of the `Cache` interface.

To extend the Blueprint IR with the new Choice, follow the following steps: 

+ In the [generators](generators) folder, create a new file for the choice.
+ In the newly created file, add a new `struct` that defines the IR node for the choice. The struct must implement the `Node` interface described in [generators/ir.go](generators/ir.go).
+ Add a new Visit method for the `Visitor` interface defined in [generators/visitor.go](generators/visitor.go). Implement the new Visit method for the `DefaultVisitor` struct defined in the same file so that all of the other visitors implement the newly extended `Visitor` interface.
+ In the choice file, add a Generator function that creates a new object of the IR Node type.
+ Register the generator function for the Choice in the IRRegistry. To do so, add a new entry to the `reg` map in the `InitIRRegistry` function defined in [generators/ir_extension.go](generators/ir_extension.go) that maps the choice to the generator function defined in the previous step.
+ The Visit method for the choice node must be overriden in the following visitors: [generators/print_visitor.go](generators/print_visitor.go) and [generators/client_collector_visitor.go](generators/client_collector_visitor.go).
+ If the choice requires a server to be deployed (eg. Redis server for the Redis choice), then different visitors need to be updated so that they can add deployment information for the choice. The visitors that need to be updated are: [generators/addr_collector_visitor.go](generators/addr_collector_visitor.go), [generators/basic_deploy_visitor.go](generators/basic_deploy_visitor.go), and [generators/main_visitor.go](generators/main_visitor.go).


### __Adding New Features__

Adding a new feature can potentially include extending Blueprint in a multitude of ways. Here we highlight some of the various ways, one may have to extend Blueprint to add one specific feature.
Note that, while we highlight all the ways in which Blueprint we believe majority of the features would require only one type of modification.

+ Adding a SourceCodeModifier
+ Adding a DeploymentModifier
+ Adding a New Server Framework
+ Adding a New Deployer Framework
+ Adding an IR Modifier (TODO)
+ Adding a new Config Option
+ Extending the Wiring Language
+ Extending the Specification Language

#### __Adding a SourceCodeModifier__

Add a SourceCodeModifier to add a feature that can behave as an interceptor for APIs on the source code side. A SourceCodeModifier modifies the behaviour of the underlying service. An example of a SourceCodeModifier would be TracingModifier that adds wraps the internal code with tracing calls.
We describe the steps needed to add a new SourceCodeModifier using a hypothetical FooModifier that prints ```Hello World!``` at the start of every API call.

1. In the [generators](generators) folder, create a new file called ```foo_modifier.go```
2. In the foo_modifier.go file,, add a ```struct``` called ```FooModifier``` that embeds a ```NoOpSourceCodeModifier``` and has a field called ```Params``` of type ```[]Parameter```. This struct will be the type generated for the IR Node for the FooModifier.
The struct should look like the following:

```go
type FooModifier struct {
    *NoOpSourceCodeModifier
    Params []Parameter
}
```

3. Implement the methods to satisfy the Node interface

```go
func (m *FooModifier) Accept(v Visitor) {
    v.VisitFooModifier(v, m)
}

func (n *FooModifier) GetNodes(nodeType string) []Node {
    var nodes []Node
    if getType(n) == nodeType {
        nodes = append(nodes, n)
    }
    for _, child := range n.Params {
        nodes = append(nodes, child.GetNodes(nodeType)...)
    }
    return nodes
}
```

4. Modify the ```Visitor``` interface to add the ```VisitFooModifier``` function prototype in [generators/visitor.go](generators/visitor.go)

```go
type Visitor interface {
    // ...
    // Other Visit-IRNode methods
    // ...
    // Modifiers
    // ... Modifier specific Visitor methods
    VisitFooModifier(v Visitor, n *FooModifier)
}
```

5. Add implementations of the ```VisitFooModifier``` method to the ```DefaultVisitor``` class and ```PrintVisitor``` class. Note that for Blueprint to build properly, adding an implementation of the method to the ```DefaultVisitor``` class is mandatory as this is the Visitor class embedded in all the other Visitor classes. Adding an implementation to the ```PrintVisitor``` class is not mandatory but recommended.

```go
func (_ *DefaultVisitor) VisitFooModifier(v Visitor, n *FooModifier) {
    for _, node := range n.Params {
        node.Accept(v)
    }
}
```

6. Register the new Modifier in the Modifier registry in the [generators/modifier.go](generators/modifier.go) file.

```go
func InitModifierRegistry(logger *log.Logger) *ModifierRegistry {
    // ...
    // Other registrations
    reg["FooModifier"] = GenerateFooModifier

    // Network Boundary Modifier registrations
    // ...

    // Server-Client Boundary Modifier registrations
    // ...
}
```

7. Register the new Modifier in the Wiring Modifier registry in the file [wiring/registry.py](wiring/registry.py) file.

```python
valid_client_modifiers = {
    # ...
    # Other modifiers
    "FooModifier",
}

valid_server_modifiers = {
    # ...
    # Other modifiers
    "FooModifier",
}
```

8. In the foo_modifier.go file, implement the methods to satisfy Modifier interface. Note that by embedding the ```NoOpSourceCodeModifier``` in our ```FooModifier``` struct, we only need to override the methods needed by this modifier.
We implement the basic ```Modifier``` methods and the methods needed to modify the behaviour at a source code level.

```go
// Basic Modifier Methods
func (m *FooModifier) GetParams() []Parameter {
    return m.Params
}

func (m *FooModifier) GetName() string {
    return "FooModifier"
}

func GenerateFooModifier(node parser.ModifierNode) Modifier {
    return &FooModifier(NewNoOpSourceCodeModifier(), get_params(node))
}

// Source-Code Modification Methods

// Method to modify the client side behaviour
func (m *FooModifier) ModifyClient(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
    // Map to hold the function bodies for each function name
    bodies := make(map[string]string)
    // Copy the API function signatures from the previous node in the client node chain
    newMethods := copyMap(prev_node.Methods)
    // Set the receiver name to be used by the methods
    receiver_name := "foo"
    // Modify each method so that it prints "Hello World!"
    for name, method := range newMethods {
        // Add any new Args or Return Values required by the previous node to the arg_list/return_list of this new method
        combineMethodInfo(&method, prev_node)
        // Generate the body of the method
        var body string
        // Add a "Hello, World" print statement 
        body += "println(\"Hello, World!\")"
        // Call the API on the next node in the client node chain
        var arg_names []string
        for _, arg := range method.Args {
            arg_names = append(arg_names, arg.Name)
        }
        body += "return " + receiver_name + ".client." + name + "(" + strings.Join(arg_names, ", ") + ")\n"
        // Add the function_name:body mapping to the map of all function bodies
        bodies[name] = body
        // Update the modified function in the mapping.
        newMethods[name] = method
    }
    // Construct the Object that will be used for generation during the Source Code generation process
    // next_node_args is left empty as FooModifier has not added any new arguments or return values to the source code Modifier
    next_node_args := []parser.ArgInfo{}
    // imports is left empty as this doesn't require any new imports
    imports := []parser.ImportInfo{}
    // Set the Name of the Object that will be generated
    name := prev_node.BaseName + "FooClient"
    return &ServiceImplInfo{Name: name, ReceiverName:receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: imports, InstanceName: prev_node.InstanceName, NextNodeMethodArgs: next_node_args}, nil
}

// Method to modify the server side behaviour
func (m *FooModifier) ModifyServer(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
    // Map to hold the function bodies for each function name
    bodies := make(map[string]string)
    // Copy the API function signatures from the previous node in the client node chain
    newMethods := copyMap(prev_node.Methods)
    // Set the receiver name to be used by the methods
    receiver_name := "foo"
    // Modify each method so that it prints "Hello World!"
    for name, method := range newMethods {
        // Add any new Args or Return Values required by the previous node to the arg_list/return_list of this new method
        combineMethodInfo(&method, prev_node)
        // Generate the body of the method
        var body string
        // Add a "Hello, World" print statement
        body += "println(\"Hello, World!\")"
        // Call the API on the next node in the server node chain
        var arg_names []string
        for _, arg := range method.Args {
            arg_names = append(arg_names, arg.Name)
        }
        body += "return " + receiver_name + ".service." + name + "(" + strings.Join(arg_names, ", ") + ")\n"
        // Add the function_name:body mapping to the map of all function bodies
        bodies[name] = body
        // Update the modified function in the mapping.
        newMethods[name] = method
    }
    // Construct the Object that will be used for generation during the Source Code generation process
    // imports is left empty as this doesn't require any new imports
    imports := []parser.ImportInfo{}
    // Set the Name of the Object that will be generated
    name := prev_node.BaseName + "FooServer"
    // Generate the constructor for this object
    constructor_name := "New" + name
    ret_args := []parser.ArgInfo{parser.GetPointerArg("", name)}
    args := []parser.ArgInfo{parser.GetPointerArg("service", prev_node.Name)}
    constructor := parser.FuncInfo{Name: constructor_name, Args: args, Return: ret_args}
    cons_body := "return &" + name + "{service: service}\n"
    bodies[constructor.Name] = cons_body
    // Set the struct fields for this object
    fields := []parser.ArgInfo(parser.GetPointerArg("service", prev_node.Name))
    return &ServiceImplInfo{Name: name, ReceiverName:receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: imports, InstanceName: prev_node.InstanceName, NextNodeMethodArgs: next_node_args, Constructors: []parser.FuncInfo{constructor}, BaseImports: prev_node.BaseImports, Fields: fields}, nil
}

// Function to generate the Constructor method for the Client-Side FooNode
// A separate function is needed as this can not be generated during the ClientNode method modification process as the next node in the Client chain is unknown at that time.
func (m *FooModifier) AddClientConstructor(node *ServiceImplInfo, next_node *ServiceImplInfo) {
    // Set the name of the constructor function
    constructor_name := "New" + node.Name
    // Set the return arguments of the constructor function
    ret_args := []parser.ArgInfo{parser.GetPointerArg("", node.Name)}
    // Set the incoming arguments of the constructor function
    args := []parser.ArgInfo{parser.GetPointArg("client", next_node.Name)}
    // Set the body of the constructor function
    body := "return &" + node.Name + "{client:client}\n"
    // Generate the Constructor info object
    constructor := parser.FuncInfo{Name: constructor_name, Args: args, Return: ret_args}
    // Add the constructor method to the node
    node.MethodBodies[constructor_name] = body
    node.Constructors = []parser.FuncInfo{constructor}
    // Add the fields in the struct for the ClientNode
    var fields []parser.ArgInfo
    fields = append(fields, parser.GetPointerArg("client", next_node.Name))
    node.Fields = fields
}
```

#### __Adding a DeploymentModifier__

Add a DeploymentModifier to add a feature that modifies the deployment behaviour of the application. We describe the steps needed to add a new DeploymentModifier using a hypothetical BarModifier that adds a new environment variable called ```BAR``` with the value ```Hello World```.

1. In the [generators](generators) folder, create a new file called ```bar_modifier.go```
2. In the bar_modifier.go file,, add a ```struct``` called ```BarModifier``` that embeds a ```NoOpDeployerModifier``` and has a field called ```Params``` of type ```[]Parameter```. This struct will be the type generated for the IR Node for the BarModifier.
The struct should look like the following:

```go
type BarModifier struct {
    *NoOpSourceCodeModifier
    Params []Parameter
}
```

3. Implement the methods to satisfy the Node interface

```go
func (m *BarModifier) Accept(v Visitor) {
    v.VisitBarModifier(v, m)
}

func (n *BarModifier) GetNodes(nodeType string) []Node {
    var nodes []Node
    if getType(n) == nodeType {
        nodes = append(nodes, n)
    }
    for _, child := range n.Params {
        nodes = append(nodes, child.GetNodes(nodeType)...)
    }
    return nodes
}
```

4. Modify the ```Visitor``` interface to add the ```VisitBarModifier``` function prototype in [generators/visitor.go](generators/visitor.go)

```go
type Visitor interface {
    // ...
    // Other Visit-IRNode methods
    // ...
    // Modifiers
    // ... Modifier specific Visitor methods
    VisitBarModifier(v Visitor, n *BarModifier)
}
```

5. Add implementations of the ```VisitBarModifier``` method to the ```DefaultVisitor``` class and ```PrintVisitor``` class. Note that for Blueprint to build properly, adding an implementation of the method to the ```DefaultVisitor``` class is mandatory as this is the Visitor class embedded in all the other Visitor classes. Adding an implementation to the ```PrintVisitor``` class is not mandatory but recommended.

```go
func (_ *DefaultVisitor) VisitBarModifier(v Visitor, n *BarModifier) {
    for _, node := range n.Params {
        node.Accept(v)
    }
}
```

6. Register the new Modifier in the Modifier registry in the [generators/modifier.go](generators/modifier.go) file.

```go
func InitModifierRegistry(logger *log.Logger) *ModifierRegistry {
    // ...
    // Other registrations
    reg["BarModifier"] = GenerateBarModifier

    // Network Boundary Modifier registrations
    // ...

    // Server-Client Boundary Modifier registrations
    // ...
}
```

7. Register the new Modifier in the Wiring Modifier registry in the file [wiring/registry.py](wiring/registry.py) file.

```python
valid_server_modifiers = {
    # ...
    # Other modifiers
    "BarModifier",
}
```

8. In the bar_modifier.go file, implement the methods to satisfy Modifier interface. Note that by embedding the ```NoOpDeployerModifier``` in our ```BarModifier``` struct, we only need to override the methods needed by this modifier.
We implement the basic ```Modifier``` methods and the methods needed to modify the behaviour of the Deployment Information.

```go
func (m *BarModifier) ModifyDeployInfo(depInfo *deploy.DeployInfo) error {
    depInfo.EnvVars["Bar"] = "Hello World"
    return nil
}
```

#### __Adding a new Server framework__

Currently Blueprint supports 2 kinds of server frameworks - rpc and http (web). These server frameworks are responsible for generating the handlers around service implementations as well as clients for correctly contacting the services.

Each new server framework must satisfy the `NetworkGenerator` interface defined in [generators/netgen/framework.go](generators/netgen/framework.go):

```go
type NetworkGenerator interface {
    // Sets the name of the application
	SetAppName(appName string)
    // Conversion Methods
    // Convert the remote data types to the format required by this Network Generator.
	ConvertRemoteTypes(remoteTypes map[string]*parser.ImplInfo) error
    // Convert the enum data types to the format required by this Network Generator
	ConvertEnumTypes(enums map[string]*parser.EnumInfo) error
    // Generate the necessary files. For grpc, this includes generating the .proto file for the application and generating the code using the Protobuf compiler for generating the low-level RPC code.
	GenerateFiles(outdir string) error
    // Source Code Generation Methods
    // Generate request de-serialization code on the server-side
	GenerateServerMethods(handler_name string, service_name string, methods map[string]parser.FuncInfo) (map[string]string, error)
    // Generate request serialization code on the client-side
	GenerateClientMethods(handler_name string, service_name string, methods map[string]parser.FuncInfo, nextNodeMethodArgs []parser.ArgInfo, nextNodeMethodReturn []parser.ArgInfo) (map[string]string, error)
    // Generate the Server-Side node for launching and starting the server + request handler
	GenerateServerConstructor(prev_handler string, service_name string, handler_name string, base_name string) (parser.FuncInfo, string, []parser.ImportInfo, []parser.ArgInfo, []parser.StructInfo)
    // Generate the Client-Side for establishing the connection to the Server
	GenerateClientConstructor(service_name string, handler_name string, base_name string) (parser.FuncInfo, string, []parser.ImportInfo, []parser.ArgInfo, []parser.StructInfo)
    // Returns the new imports to be added for the constructed Objects by this Network Generator
	GetImports(hasUserDefinedObjs bool) []parser.ImportInfo
    // Returns the list of dependencies to enable generation with this Network Generator
	GetRequirements() []parser.RequireInfo
}
```

#### __Adding a new Deployment Framework__

```go
type DeployerGenerator interface {
    // Function to add a new User-Defined instance to the current deployment configuration
	AddService(name string, depInfo *DeployInfo)
    // Function to add a Choice instance for a Component in-use to the current deployment configuration
	AddChoice(name string, depInfo *DeployInfo)
    // Function that generates the relevant configuration files for deployment
	GenerateConfigFiles(out_dir string) error
}
```

#### __Adding an IR Modifier__

Documentation coming soon....

#### __Adding a new config option__

If you wish to add a new configuration option then please follow the following steps:

1. Update the config parser located in the file [parser/config.go](parsing/config.go)
2. Update the [main.go](main.go) so that the necessary config options are passed down to the relevant visitor passes.

#### __Extending the Wiring Language__

If you wish to extend the syntax of the Wiring language then please follow the following steps:

1. Update the wiring parser located in the file [wiring/parser.py](wiring/parser.py)
2. Update the data models if needed. The data models are located in the [wiring/dataModels.py](wiring/dataModels.py) file.
3. Update the partial evaluator if needed. The partial evaluator for modifiers is located in the file [wiring/eval.py](wiring/eval.py). Modifications would be needed if the syntax changes to the wiring parser impact how modifiers are defined.

#### __Extending the Specification Language__

If you wish to extend the syntax of the Specification language, for e.g. to support features from future versions of Go, then please follow the following steps:

1. Update the spec parser located in the file [parser/spec.go](parser/spec.go)
2. IF adding a new type, then update the [parser/types.go](parser/types.go)