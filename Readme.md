Lava Proxy server:
server.go file, run the command "go run server.go" in terminal.

To test it, make sure grpcurl is installed.
The following commands can be run in terminal:
grpcurl -plaintext localhost:50051 describe cosmos.base.tendermint.v1beta1.Service

State Tracker:
After running the lava proxy server state tracker can be used.
To run state tracker for the latest 5 blocks run the command "go run state_tracker.go" in terminal.
State tracker will print to a file the height and hash values of the 5 latest blocks.