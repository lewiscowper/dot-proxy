# dot-proxy

DNS to DNS over TLS proxy

---

## Commit History

- This commit takes in a basic dns server with https://godoc.org/github.com/miekg/dns. It responds to queries from kdig over tcp, (and udp is but a goroutine away). Next steps will be taking what I have now and forwarding the requests over tls to an external dns server.
- Ran gofmt because I realised I had my indents all muddled up. (Still setting up my personal laptop for go development).
- Initial commit with README, and a basic net.Listen and net.Dial that forwards packets, as verified by netcat listening on port 2411, and broadcasting packets from port 2410. Mostly just a way of playing around with new bits of go's standard library that I haven't used. Also experimented with pushing packets into it from kdig, to see what data might get dumped along the way. Lots of research and exploration at this stage, it is still nowhere near a DNS - DoT proxy, but it's been interesting pushing packets around. (Although only in one direction).
