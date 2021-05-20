# mqtt server

## Status
CONNECT, SUBSCRIBE, PUBLISH (with any qos) seems work fine.
retain messages and sending will under development
also persistent still doesn't work (temporary;)

## Purpose
This is my implementation of mqtt server for Maja Suite project. Mqtt server will be central queue for all messages 
from connected devices and hub.

Drivers (sonoff, xiaomi or zigbee) will connect to mqtt server at start and push all messages to mqtt. Hub will receive,
manage and push messages back. Interconnect beetween devices different types will be managed by mqtt server.

## Save data on restart
Server use sqlite database to store username/password as well as will/retain messages. So restarts should be clear.

## Code
I write code as simple as possible, so it should (I hope) supported very easy. May be somewhere it looks not very 
professional, in this case kindly drop me message or pull request (if you can).

Surelly some errors may exists. Use it at your own risk ;)

## Notice
Specification: https://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html

Author: Eugene Chertikhin <e.chertikhin@crestwavetech.com>

Licensed under GNU GPL.

## Problems on mqtt 3.1.1 specification (should keep it in mind)

### No errors manage
I.e. can't send error to publish command in case of wrong topic or illegal payload.

