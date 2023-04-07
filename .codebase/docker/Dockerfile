FROM istio/proxyv2:1.16.3 as proxyv2
FROM hub.byted.org/istio-testing/build-tools:release-1.16-fb2454846dcbf0275675ce7d7e22e2602ff1aa38
COPY --from=proxyv2 /usr/local/bin/envoy /usr/local/bin/envoy