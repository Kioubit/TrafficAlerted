# TrafficAlerted

## Monitor transiting traffic. Per-IP alerts on high traffic volume. Network scanning alerts.

![Network Flow Chart](screenshot.png?raw=true)


- Monitors traffic on all interfaces
- Monitors traffic based on source ip and destination ip pairs, generate network flow chart
- Alerts on bandwidth exceeded
- Detects port/network scanning activity (Via the number of different IPs contacted and Number of ports contacted)
- Does not require root access. Only required capability: CAP_NET_RAW
- Does not require any configuration

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

Exclude interfaces from monitoring by creating a file named "excluded-interfaces" in the working directory of the program and adding each interface to be excluded on a new line.

### Building
#### Manual

You will need to have GO installed on your system. Then run `make release` and find the binary in the `bin/` directory.

***

### Module Documentation

#### mod_httpAPI
Provides the following http API endpoints on port `8698`:

- `/active`

It also provides the following user interface:
- `/dashboard`
