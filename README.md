# dot-proxy

DNS to DNS over TLS proxy

---

## Startup

Locally, run the following commands:

```
docker build -t dot-proxy:latest . &&\
docker run --rm -it -p 2410:53/tcp -p 2410:53/udp -e DOTPROXY_UPSTREAM_HOST=1.1.1.1 -e DOTPROXY_UPSTREAM_PORT=853 -e DOTPROXY_LISTEN_PORT=53 dot-proxy
```

From there, you can verify that DNS queries work with the following commands:

```
kdig -d @0.0.0.0:2410 lewiscowper.com
kdig +tcp -d @0.0.0.0:2410 lewiscowper.com
```

## Security Concerns

There are some security concerns that should be evaluated when running this service.

As a kick off, DNS over TLS is a hop-to-hop encrypted protocol, and while it does provide some obfuscation of the actual queries being made, it doesn't keep your DNS queries safe from the prying eyes of the administrators of the company running your chosen upstream DNS resolver. As opposed to a (non-existent at this stage as far as I can tell) different protocol that offered end to end encryption. Based on my understanding of how DNS works, I'm not sure how feasible that would be, if ever, as the resolver needs to know the address being requested, so it can return the appropriate records to you. Perhaps there's some kind of asymmetric cryptography thing that might allow that, but that's very much not on the horizon as far as I can tell. Furthermore, just on the protocol end, there are two further (but thankfully much shorter) issues, one is that frequently a resolver that offers DNS over TLS is actually going to send your request to an authoritative nameserver unencrypted anyway, so the communication between you and your resolver is encrypted, but that encryption doesn't continue on. Building on the two previous points, yes DNS over TLS could be something useful, but it has to be implemented at each hop.

So, back to less broad protocol strokes, in this case, where we have a proxy service running alongside other services, and we want to route all DNS traffic to a DoT proxy before it exits the cluster/server/environment etc, in this implementation, we'd want to not log anything other than response time (and that we could really push out to prometheus or other monitoring data ingest tools), because otherwise if the proxy gets compromised, every requested hostname is in the logs. (I found it much easier to debug as I was building to have the query in there, but in case privacy is more important than knowing which query broke things, then I'd definitely make that change.

## How I'd use this proxy in a microservice architecture

In order to get rid of traffic going between containers to do DNS lookups, in a Kubernetes environment, I'd take advantage of "sidecar" containers, and run an instance alongside each service deployment. This would mean pod-internal traffic was unencrypted (over the regular DNS TCP or UDP ports) but as it left the pod to go to the upstream host, it would be encrypted with TLS. Done with appropriate scaling (as in not a large amount of CPU/RAM requests, the container takes very little to run), I think that strikes a useful balance between accessibility to the services (they set up their DNS resolver to a container local address, and they're done), and security. This also has the added benefit that only services that needed to make external DNS calls would need the proxy, and you could entirely restrict external traffic going out of the cluster on port 53.

There's also an argument to be made for running (at least) one instance per physical machine in a DaemonSet. This might be more useful in a situation where most services need external DNS, and the goal was to encrypt all DNS traffic in-flight. This would be less of a hassle for application developers having to add sidecar containers to run DNS, and would be simpler to deploy and run in terms of raw YAML quantity. But it would mean that if the cluster network was breached, all DNS traffic could be sniffed inside the cluster. (Although to be honest, if the cluster network was breached, leaking DNS queries is probably quite low down on the list of things to worry about).

## Future improvements

I enumerated some of these in the checklist below, and alluded to one of them above, so I'll keep it somewhat brief.

- Prometheus metrics instead of logging for response time.
  I did some basic "load" testing, by opening 5 terminals and running kdig in a loop every second, and didn't notice the request time getting huge, but it'd be nice to both scale that load testing up, and try and push the service to it's limits. (Although, based on my testing, it's far more likely that the container would need to restart due to an interrupted connection to cloudflare).
- On upstream connection interruption, reconnect.
  This one is the most annoying, as it definitely doesn't reflect well on me. Given more time that would be first thing that I'd be looking to fix. I believe it can be triggered by issuing multiple simultaneous requests, and there's also an i/o timeout that may well be cloudflare doing some rate limiting. Digging into that rabbit hole, although I still have more than a few hours, (although limited by my schedule and sleep), I think it's something I'm okay with leaving as an unimplemented bugfix, as handling multiple connections does work, it's just that as each request takes around 20ms to run (in Docker, on my machine, your speeds may vary), it's quite difficult to force coincidental requests.
- Add a helm template or another way of integrating into kubernetes.
  This one I'm not too fussed about, especially if we followed the sidecar container idea, then the service's helm chart would pull the image, set some environment variables, and not have anything more to do.
- Multiple upstreams.
  Allowing multiple upstreams would be a real delight, it would mean thinking more about how to handle the configuration, as environment variables are notoriously difficult to co-erce into arrays at the best of times. But more generally, having a round robin or other distribution method for DNS queries across a range of DoT providers would generally be great, and having the proxy try a new connection if a query fails, or a provider's connection timed out would be a boon for the reliability of the proxy generally, and by extension the reliability of DNS queries across all services that need them.
- Adding tests.
  I've been testing heavily with `kdig` as that allows me to specify testing with the TCP protocol, and was a real help early on in the project when I was still finding my bearings, and figuring out things that worked, and things that did not work. Plus the implementation is so comparably small (under 100 lines including whitespace etc), that having unit tests seems somewhat overkill, but having them along with some basic integration tests that determine whether (for example) cloudflare.com responds to a query appropriately (obviously relying on the specific returned IP in the record wouldn't work, but generally that the request succeeded would be enough), would be a useful start to being comfortable pushing this to a production environment.

## Checklist

- [x] Package into Dockerfile
- [x] Expose on port 53
- [ ] Add prometheus metrics (just for fun)
- [ ] Make new connection to upstream (cloudflare for now) on i/o timeout
- [x] Remove essentially anything hard coded and move into configuration.
- Documentation
  - [x] What are the security concerns for this kind of service?
  - [x] Considering a microservice architecture, how would you see the dns to dns-over-tls proxy used?
  - [x] What other improvements do you think would be interesting to add to this project?
- [ ] Any other stretch goals. (Helm template? Sidecar to busybox etc? Round robin/other selection method across multiple upstreams? Add tests)

## Commit History (latest first)

- Added Dockerfile, testable via `kdig +tcp -d @0.0.0.0:2410 lewiscowper.com` after running the startup command specified above. Added configuration with [envconfig](https://github.com/kelseyhightower/envconfig). Updated documentation to finish off the challenge.
- Added UDP server, as it was literally two lines.
- Looks like we have a DNS to DNS over TLS proxy, after another few hours today. I'll add a udp server too, just to make my kdig query shorter, and tomorrow can be reserved for packaging and documenting usage and how to test. What I'm still not too sure about is how exactly to verify that the connection is being made, aside from the fact that the previous commit wouldn't respond with anything useful nameserver or DNS record wise, and this commit does. That means I'm definitely connecting to cloudflare's DNS servers generally, but I'm not 100% sure how to validate that it's going over DoT, other than I'm hitting 1.1.1.1:853. Something to think about in the documentation I guess
- This commit takes in a basic DNS server with https://godoc.org/github.com/miekg/dns. It responds (with an empty body) to queries from kdig over tcp, (and udp is but a goroutine away). Next steps will be taking what I have now and forwarding the requests over tls to an external dns server.
- Ran gofmt because I realised I had my indents all muddled up. (Still setting up my personal laptop for go development).
- Initial commit with README, and a basic net.Listen and net.Dial that forwards packets, as verified by netcat listening on port 2411, and broadcasting packets from port 2410. Mostly just a way of playing around with new bits of go's standard library that I haven't used. Also experimented with pushing packets into it from kdig, to see what data might get dumped along the way. Lots of research and exploration at this stage, it is still nowhere near a DNS - DoT proxy, but it's been interesting pushing packets around. (Although only in one direction).
