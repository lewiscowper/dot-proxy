# dot-proxy

DNS to DNS over TLS proxy

---

## Final checklist

- [ ] Package into Dockerfile
- [ ] Expose on port 53
- [ ] Add prometheus metrics (just for fun)
- [ ] Make new connection to upstream (cloudflare for now) on i/o timeout
- [ ] Remove essentially anything hard coded and move into configuration.
- Documentation
  - [ ] What are the security concerns for this kind of service?
  - [ ] Considering a microservice architecture, how would you see the dns to dns-over-tls proxy used?
  - [ ] What other improvements do you think would be interesting to add to this project?
- [ ] Any other stretch goals. (Helm template? Sidecar to busybox etc? Round robin/other selection method across multiple upstreams? Add tests(?!))

## Commit History (latest first)

- Added UDP server, as it was literally two lines.
- Looks like we have a DNS to DNS over TLS proxy, after another few hours today. I'll add a udp server too, just to make my kdig query shorter, and tomorrow can be reserved for packaging and documenting usage and how to test. What I'm still not too sure about is how exactly to verify that the connection is being made, aside from the fact that the previous commit wouldn't respond with anything useful nameserver or DNS record wise, and this commit does. That means I'm definitely connecting to cloudflare's DNS servers generally, but I'm not 100% sure how to validate that it's going over DoT, other than I'm hitting 1.1.1.1:853. Something to think about in the documentation I guess
- This commit takes in a basic DNS server with https://godoc.org/github.com/miekg/dns. It responds (with an empty body) to queries from kdig over tcp, (and udp is but a goroutine away). Next steps will be taking what I have now and forwarding the requests over tls to an external dns server.
- Ran gofmt because I realised I had my indents all muddled up. (Still setting up my personal laptop for go development).
- Initial commit with README, and a basic net.Listen and net.Dial that forwards packets, as verified by netcat listening on port 2411, and broadcasting packets from port 2410. Mostly just a way of playing around with new bits of go's standard library that I haven't used. Also experimented with pushing packets into it from kdig, to see what data might get dumped along the way. Lots of research and exploration at this stage, it is still nowhere near a DNS - DoT proxy, but it's been interesting pushing packets around. (Although only in one direction).
