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
    wget https://raw.githubusercontent.com/Kioubit/pndpd/master/TrafficAlerted.conf -P /etc/TrafficAlerted/
    ````
5) Edit the config at ``/etc/pndpd/TrafficAlerted.conf`` and then start the service using ``service TrafficAlerted start``



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
