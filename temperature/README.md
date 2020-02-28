# Temperature script

Expose DS18B20 sensors as HTTP service.

## Setup
Connect one or more DS18B20 sensors to GPIO 4 and configure your raspberry pi according to https://github.com/yryz/ds18b20.

To run at boot:
1. Copy etc_systemd_system_temperature.service to /etc/systemd/system/temperature.service
2. `sudo systemctl enable temperature.service`

## Calling:
GET http://{your pi ip}:8002/temperatures/all. All sensors detected by the library will be returned. Temperatures are returned in Fahrenheit.

Example:

```
{"sensors":{"28-030897946183":"71.38"}}
```

## Home Assistant integration
### Config
Edit configuration.yml to add the sensors and then reload the config.

```
sensors:
  - platform: rest
    name: coopTemp
    resource: http://192.168.0.202:8002/temperatures/all
    value_template: '{{ value_json.senosrs["28-03029794593f"] }}'
```

The last four lines can be repeated for different sensor IDs, if you have more than one sensor.

### UI
To display the temperature:
1. Configure UI
2. Add manual card
3. Edit the following, accodingly
```
cards:
  - entity: sensor.cooptemp
    name: Coop Temperature
    theme: Backend-selected
    type: sensor
    unit: °F
type: vertical-stack
```
