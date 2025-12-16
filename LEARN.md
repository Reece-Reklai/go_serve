# Go Learnings + System Defintions + HTTP Fundamentals
##  Go Learnings
> go build -o out && ./out
- **go build** command compiles the application into an executable to run it in simplest terms.
    - go build -o (-o force the build to write the resulting executable or object to the named output file or directory).
> reference: go help build
##  System Defintions
- **Process** is an active, independent instance of a program with its own memory space.
- **Thread** is a lightweight unit of execution within a process that shares the process's memory and resources.
> analogy = the company is a process while threads are its employees
- **Stack** - local variables and functions calls are stored.
- **Heap** - user allocated memory is stored.
## HTTP Fundamentals
- **RESTAPI** design
    - Resource Plural naming for api endpoints (example: videos instead of video)
        - GET /api/videos (Get all videos) **Get all videos**
        - GET /api/videos/{id} **Gets individual**
        - POST /api/videos **Server responsible provides the new id in its response**
        - PUT /api/videos/{id} **Update a video**
        - DELETE /api/video/{id} **Delete a video**
- **JWT** design
    - Client Loing
    - JWT + user_id created and sent back to client
    - Client sends JWT in all future requests
    - On every authentication request: server validates JWT
- **Token** concepts
    - Access Token is what gives a user the ability to access their resource over a webserver without having to login every time
    - Refresh Token is used to give user the ability to obtain new access tokens
##### https://restfulapi.net/
- **Socket** is the fundamental abstraction for network communication.
    - Acts as an interface between application and transport layer (OSI model).
    - "Special" type of file that supports read and write.
    - TCP (connection) or UDP (connectionless) socket examples.
> analogy block
1) telephone (socket) is the medium in which the client (your phone) dials server's address and port (phone number and person you want to talk to) to establish a connection and exchange information.
2) IP address: street address of a building (correct computer on the network), Port: apartment number within the building (identify specific application or service to connect to).
3) client (person making the call to a service) and server (the person/system answer call to provide the service) while listening for incoming connection.
> analogy block
- **File Descriptor** is an integer value used by the operating systems to identify manage open files.
    - Linux/Unix Environments sees everything to be a file (directories, files, devices, network connections).
    - Starts with non-negative integer (0 stdin, 1 stdout, 2 stderr).
    - Within OS kernels, file descriptor = index pointing to a file table entry.
> analogy = file descriptor is a car key to a specific car
>> file table entry = kernel data structure stores info about currently open files and I/O resources
### Roadmap between socket and file descriptors
1) os allocates file descriptor.
2) the descriptor can be used for subsequent network operations(write, read, close, etc).
3) application interacts with socket through file descriptors.
### TCP Connection Establishment
1) socket file descriptor created.
2) server binds socket to an address and port.
3) socket listens to incoming connection request.
4) server accepts a connection.
5) client connection CONNECTS.
6) Data transmission occurs (read/write data over the connection).
7) closes connection.
### Extra Info
1) Readiness endpoints are commonly used by external systems to check if our server is ready to receive traffic
2) Passwords - store passwords not as plain text and pass strength matters
    - Hashing takes string as input and returns another string entirely (which is commonly refered to as a hash)
# Terminal Linux Commands
#### sudo lsof -i tcp:8080
- super user check list of open files including those of TCP ports 8080
> Reference: lsof --help
#### sudo kill -9 <PID>
- super user force kills(-9) on selected process
> Reference help kill
# Credit
##### Boot Dev Courses
##### https://dev.to/leapcell/behind-the-scenes-tcp-connections-in-gos-nethttp-46ek
##### https://unicminds.com/program-vs-process-vs-thread/
