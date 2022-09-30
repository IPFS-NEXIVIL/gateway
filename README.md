1.

- `http://localhost:8001/ipfs/relay-node`

make relay host

2.

Notify the listening node of relay host

```
ipfs swarm connect <relay host address>
```

3.

- `http://localhost:8001/ipfs/connect-to-local-ipfs`

`request`

```
{
    "relay_host": "relay host address",
    "dial_id": "listening node peer id"
}
```

4.  `Result`

```
Okay, no connection from h1 to h3:  no addresses
Just as we suspected
end
```

---

<img width="909" alt="스크린샷 2022-09-30 오후 2 40 36" src="https://user-images.githubusercontent.com/73830753/193198271-5da8bdc4-051e-4891-94bc-78fb6b7139e7.png">
