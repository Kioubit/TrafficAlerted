# TrafficAlerted

## Monitor transiting traffic. Per-IP alerts on high traffic volume. Network scanning alerts.

![Network Flow Chart](screenshot.png?raw=true)

## Features

- Monitors traffic on all interfaces
- Monitors traffic based on source ip and destination ip pairs, generates a network flow chart
- Alerts on per-ip bandwidth exceeded
- Detects port/network scanning activity (Via the number of different IPs contacted and the number of ports contacted)
- Does not require root access. Only required capability: CAP_NET_RAW
- Minimal configuration

## Installing & Updating

1) Download the latest release from the [releases page](https://github.com/Kioubit/TrafficAlerted/releases) and move the binary to the ``/usr/local/bin/`` directory under the filename ``TrafficAlerted``.
2) Allow executing the file by running ``chmod +x /usr/local/bin/TrafficAlerted``
3) **For systemd users:** Install the service unit file
    ```` 
    wget https://raw.githubusercontent.com/Kioubit/TrafficAlerted/master/TrafficAlerted.service -P /etc/systemd/system/
    systemctl enable TrafficAlerted.service
    ```` 
4) Download and install the config file
    ```` 
    mkdir -p /etc/TrafficAlerted
    wget https://raw.githubusercontent.com/Kioubit/TrafficAlerted/master/TrafficAlerted.conf -P /etc/TrafficAlerted/
    ````
5) Edit the config at ``/etc/TrafficAlerted/TrafficAlerted.conf`` and then start the service using ``service TrafficAlerted start``



### Building
#### Manual

You will need to have GO installed on your system. Then run `make release` and find the binary in the `bin/` directory.

***

### Module Documentation

#### mod_httpAPI
Provides the following http API endpoints on port `8698`:

- `/active`
- `/capabilities`

It also provides the following user interface:
- `/dashboard`

## Example configuration file
      # Configuration file for TrafficAlerted
      
      # General configuration
      # --------------------------------------------------
      
      # The amount of time in seconds for which "ByteLimit" applies
      TrafficInterval: 170
      # The amount of bytes to trigger an alert for. Either specify bytes or add the suffixes "K","M" or "G"
      ByteLimit: 1G
      # If set to 'true', the program does not keep destination IP information.
      # If set to 'true', Port scanning detection based on the number of contacted IPs is then disabled as well.
      # If set to 'true', Detailed information about the traffic to the destination IPs is also unavailable.
      NoDestinations: false
      # Minimum amount of bytes for an IP to appear on the network chart
      ChartMinBytes: 5M
      # Exclude interfaces from monitoring or leave empty to monitor all
      ExcludedInterfaces: "lo","dummy0"
      
      
      # Network and Port scanning detection
      # --------------------------------------------------
      # Enable analyzing the amount of IPs a given host contacts
      ContactedIPsAnalyze: true
      # If a source IP contacts more than this amount of different IPs within "TrafficInterval" scanning activity is detected.
      NumberContactedIPs: 15
      
      # Whether to analyze destination ports for scanning activity (Requires more system resources). (true/false)
      AnalyzePorts: false
      # Only applicable if AnalyzePorts is true.
      # If a source IP contacts more than this amount of different ports within "PortInterval", then scanning activity is detected.
      NumContactedPorts: 100
      # Only applicable if "AnalyzePorts" is true. The amount of time in seconds for which "NumContactedPorts" applies
      PortInterval: 600
