# domsub

Confirm pubsub failure/operational domain.

* S1: Verify receive method extends message lease and prevents visibility to other consumers.
* S2: Verify unacked messages are visible after the lease expiry.

```
    +  +------------------------+
    |  |                        |
    |  | Publish Message(s)     |
    |  |                        |
    |  +------------------------+
    |
    |                            +---------------------------------------------------------+
    |                            |                      |                 |                |
 S1 |                            | 1 Receive Message(s) | Hold Message(s) | Ack Message(s) |
    |                            |                      |                 |                |
    |                            +---------------------------------------------------------+
    |
    |                                                   +----------------------------------+
    |                                                   |                                  |
    |                                                   | 2 Receive Loop Started           |
    |                                                   |                                  |
    +                                                   +----------------------------------+

    +------------------------------------------------- t ----------------------------------------------->

    +  +------------------------+                                                 +-----------------+
    |  |                        |                                                 |                 |
    |  | Publish Message(s)     |                                                 | Lease Expiry    |
    |  |                        |                                                 |                 |
    |  +------------------------+                                                 +-----------------+
    |
    |                            +---------------------------------+
    |                            |                      |          |
 S1 |                            | 1 Receive Message(s) | Die Hard |
    |                            |                      |          |
    |                            +---------------------------------+
    |
    |                                                   +-----------------------------------------------+
    |                                                   |                        |                      |
    |                                                   | 2 Receive Loop Started | Receive Message(s)   |
    |                                                   |                        |                      |
    +                                                   +-----------------------------------------------+
```

