# Door
Based on the wiring design and general idea in https://github.com/ericescobar/Chicken_Door. 

## Setup
Clone and modify .env as needed.

### Discord integration
To send discord notifications on open/close/stuck events, [create a discord webhook](https://support.discordapp.com/hc/en-us/articles/228383668-Intro-to-Webhooks) and paste the generated URL in the .env file.

### Automatic opening/closing
To automatically open/close the door at fixed times, set the appropriate cron values in .env. The go process must be restarted for updates to these values. Leave these blank if you only want to control your door manually.

Example, opening at 6:30am everyday and closing at 8pm everyday:
```
OPEN_CRON=30 6 * * *
CLOSE_CRON=0 20 * * *
```

## Web API
Three endpoints are provided. No security is in place, because I am lazy.

### Get current door status
The current door status can be obtained by GETting http://{your rpi ip}:8080/status.

The return is simply text. Possible values are the following:
* "Open"
* "Closed"
* "Opening (for x seconds)"
* "Closing (for x seconds)"
* "Stuck"

### Manually opening door
To manually open the door, POST to http://{your rpi ip}:8080/open.

### Manually closing door
To manually close the door, POST to http://{your rpi ip}:8080/close.
