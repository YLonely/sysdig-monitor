# container status json format

```json
    {
        "id": "container-id",                //container id
        "name": "container-name",            //container name
        "individual_calls": {    //different types of system call
            "name1": {
                "calls": 123,    // total invoke times
                "total_time" : 456 // total latency in nanoseconds
            },
            "name2":{}
        },
        "total_calls": 123,      // total system calls
        "io_calls_more_than_1ms": [
            {
                "file_name": "/home/test.txt",
                "latency": 123  // also nanoseconds
            },
            {

            }
        ],
        "io_calls_more_than_10ms":[],
        "io_calls_more_than_100ms":[],
        "file_total_read_in": 222, //Total bytes the container read from file system
        "file_total_write_out": 222,
        "net_total_read_in": 333,  //Total bytes the container read from network
        "net_total_write_out": 333,
        "active_connections": [
            {
                "source_ip": ,
                "dest_ip": ,
                "source_port": ,
                "dest_port": ,
                "type": "ipv4",
                "write_out": ,
                "read_in": 
            },
            {

            }
        ],
        "accessed_layers": [
            {
                "dir": "/var/lib/docker/overlay2/40f635eb604ff721234ddd0c10433377dd487419b85a1ed0c292d697b01a64a7/diff",
                "accessed_files": {
                    "/etc/bash.bashrc": {
                        "write_out": ,
                        "read_in": 
                    },
                    "...": {

                    }
                },
                "layer_write_out": ,
                "layer_read_in": 
            },
            {
                
            }
        ]
    }
```