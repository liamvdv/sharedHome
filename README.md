# sharedHome
<img src="./assets/Embedded Logo.png" width="100%">

This project is still very much work in progress. More information will follow when the software is fully written.
- [x] `vfs` for creating, loading, storing and working on the virtual filesystem representation
- [x] `config` for loading all the configuration
- [x] `stream` for chunked encryption
- [x] `osx` as a filesystem abstraction for explicit dependencies and thus testability
- [x] `backend` for storage service interface
- [ ] `backend/drive` as a storage service implementation
- [ ] `remote` package as a common wrapper around backend. handles encryption with stream.
- [ ] `core` implements the comparsion alogorithm and task execution
- [ ] `cmd` implements the commandline interface
- [ ] `signal` interface for SIGTERM and SIGINT handeling
- [ ] rething `errors`


## Goals
From the user perspective, this project has three main goals.
- First, it should have **cross-platform support** for synchronization.
- Secondly, it should employ the **zero-trust** principle. The file content and metadata is hidden from the storage provider. Only the client has the keys to read the plaintext of these files.
- Thirdly, a **modular backend** so that every storage provider can be used. I'm currently planning to start with Google Drive support.

From the technical / software architecture design perspective, the main goal of this software is to be
- **fully tested** for confident use.

## Comparsion to rsync
rsync is valued for its speed. That speed is possible because rsync compares unencrypted files and sends a operation sequence to the server, which tells it how to correct the delta. The algorithm used for that is very impressive. As some of you might have already suspected, this does not work for us. There are two main reasons for that. First, the server only knows the encrypted version of that file. That means we need to encrypt our file and then compare the differences. Based on how encryption work, small changes in the content lead to a totally different file, thus this kind of delta-correction does not work for us.
