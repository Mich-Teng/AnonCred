# What's AnonCred

AnonCred is a tool which makes user actions anonymous at anytime and anywhere in server clusters, while the clusters still maintain the ability to identify user actions based on its identity. A typical use case of it is anonymous voting. 

## How to compile and run AnonCred

This is what you need to do to compile and run AnonCred:

1.  Run `git clone https://github.com/Mich-Teng/AnonCred.git` to download source code and binary file on github. 

2.  modify `config/local.properties` to config local port.      
    modify `config/conn.properties` to config coordinator's ip and port. (for client and server only)

3.  Run `go run src/Coordinator.go` to start coordinator.       
    Run `go run src/DissentClient.go` to start client.     
    Run `go run src/AnonServerLauncher.go` to start server.       

4.  Type `msg <msg_text>` to broadcast all the messages to clients or `vote <msg_id> <+-1>` to vote towards a specific message. (For client only)

        
**Note**      
Launch coordinator first. And then launch your server and it will be automatically registered to the coordinator based on configuration. After all the servers needed are launched, type ‘ok’ in coordinator daemon to finish the server configuration. After that, you can launch client at anytime you want, but can no longer add server into the cluster.



## A simple demo

### scenario:
One coordinator, two servers and two clients.    

#### coordinator:    
tiger.zoo.cs.yale.edu:12345
servers: scorpion.zoo.cs.yale.edu:12345,  frog.zoo.cs.yale.edu:12345
#### clients:   
python.zoo.cs.yale.edu:12345,  viper.zoo.cs.yale.edu:12345

1.  Launch servers and coordinator.  
![](https://drive.google.com/file/d/0BxQd2CmLMHMcVnhMUE9ZQ3cwLVk/view?usp=sharing)
2.  Type ok to finish registration and now the coordinator starts to launch each phase.

3.  Launch the clients, sending message and voting. 





## System design

The goal of design is to make the system architecture general enough in production and improve its scalability. However, the linear model is a constraint here.


The system consists of three components: Client, coordinator and Server. The coordinator(which actually serves as a coordinator) merely serves as a coordinator and topology manager, which accepts the request from clients and deliver it to the corresponding starting server based on the phases. And it also manages to broadcast computation result to servers, verify the identity of client and periodically start each round.


The client is the program for user to send message and vote. When a new client registers, it needs to wait until next round begins to function.     


The Server is the core of system. It manages to perform encryption and decryption computations and also is responsible for broadcasting data to all the clients. (Or generally speaking, achieve any functions intreated with client-side needed by this system)


## Feedback
Questions about the system can be directly sent to MichZc.Teng@gmail.com
