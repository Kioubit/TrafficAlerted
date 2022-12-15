# TrafficAlerted

## Analyze traffic per IP address. Alert on bandwidth limit exceeded or unusual traffic.

- Monitor traffic on all interfaces
- Monitor traffic based on [source ip, destination ip] pairs
- Monitor traffic based on [source ip]
- Alert on bandwidth exceeded
- Detect port/network scanning activity (Via number of different IPs contacted and Number of ports contacted)
- No root access required. Capability: CAP_NET_RAW
- No configuration needed

### Commandline arguments
The following commandline arguments are required:

    Usage: <TrafficInterval> <ByteLimit> <NoDestinations> <NumberContactedIPs> <AnalyzePorts> <NumContactedPorts> <PortInterval>

- `TrafficInterval`: The amount of time in seconds for which `ByteLimit` applies
- `ByteLimit`: The amount of bytes to trigger an alert for
- `NoDestinations`: When 'true', the program does not keep destination IP information. Port scanning detection based on the number of contacted IPs is then disabled. Detailed information about the traffic to the destination IPs is also unavailable with this option set to 'true'
- `NumberContactedIPs`: If a source IP contacts more than this amount of different IPs within `TrafficInterval` scanning activity is detected.  
- `AnalyzePorts`: Whether to analyze destination ports for scanning activity (Requires more system resources). (true/false)
- `NumContactedPorts`: Only applicable if `AnalyzePorts` is true.  If a source IP contacts more than this amount of different ports within `PortInterval` scanning activity is detected.
- `PortInterval`: Only applicable if `AnalyzePorts` is true. The amount of time in seconds for which `NumContactedPorts` applies


### Building
#### Manual

You will need to have GO installed on your system. Then run `make release` and find the binary in the `bin/` directory.

#### Docker

Clone this repository and run `docker build .` to generate a docker image.


***

### Module Documentation

#### mod_log
Simple logger to STDOUT for events.

#### mod_httpAPI
Provides the following http API endpoints on port `8698`:

- `/active`

It also provides the following user interface:
- `/dashboard`
