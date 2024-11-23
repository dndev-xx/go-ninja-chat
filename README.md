[![CC BY-NC-SA 4.0][cc-by-nc-sa-shield]][cc-by-nc-sa]

[cc-by-nc-sa]: http://creativecommons.org/licenses/by-nc-sa/4.0/
[cc-by-nc-sa-shield]: https://img.shields.io/badge/License-CC%20BY--NC--SA%204.0-lightgrey.svg

# Bank Support Chat System

## Structure project
```
.
├── api
├── cmd
│   ├── chat-service
│   ├── gen-types
│   ├── ui-client
│   │   └── static
│   │       ├── app
│   │       └── lib
│   └── ui-manager
│       └── static
│           ├── app
│           └── lib
├── configs
├── deploy
│   └── local
├── docs
├── gorules
├── internal
│   ├── buildinfo
│   ├── clients
│   │   └── keycloak
│   ├── config
│   ├── cursor
│   ├── errors
│   ├── logger
│   ├── middlewares
│   │   └── mocks
│   ├── repositories
│   │   ├── chats
│   │   ├── jobs
│   │   ├── messages
│   │   └── problems
│   ├── server
│   │   └── errhandler
│   ├── server-client
│   │   ├── errhandler
│   │   ├── events
│   │   └── v1
│   │       └── mocks
│   ├── server-debug
│   ├── server-manager
│   │   ├── errhandler
│   │   ├── events
│   │   └── v1
│   │       └── mocks
│   ├── services
│   │   ├── afc-verdicts-processor
│   │   │   └── mocks
│   │   ├── event-stream
│   │   │   └── in-mem
│   │   ├── manager-load
│   │   │   └── mocks
│   │   ├── manager-pool
│   │   │   └── in-mem
│   │   ├── manager-scheduler
│   │   ├── msg-producer
│   │   └── outbox
│   │       └── jobs
│   ├── store
│   │   ├── chat
│   │   ├── enttest
│   │   ├── failedjob
│   │   ├── hook
│   │   ├── job
│   │   ├── message
│   │   ├── migrate
│   │   ├── predicate
│   │   ├── problem
│   │   ├── runtime
│   │   ├── schema
│   │   └── templates
│   ├── testingh
│   ├── tracing
│   ├── types
│   ├── usecases
│   │   ├── client
│   │   │   ├── get-history
│   │   │   └── send-message
│   │   └── manager
│   │       ├── can-receive-problems
│   │       ├── free-hands-signal
│   │       ├── get-chat-history
│   │       ├── get-chats
│   │       ├── resolve-problem
│   │       └── send-message
│   ├── validator
│   └── websocket-stream
├── pkg
│   └── pointer
└── tests
    └── e2e
        ├── api
        │   ├── client
        │   └── manager
        ├── client-chat
        ├── manager-workspace
        └── ws-stream
```