External blackbox for ardupilot.
---

I'm using Raspberry Pi Zero 2W + SIM7600G-H 4G HAT (waveshare) to remote control my ardupilot fixed wing.
This app protects me from losing my drone and also sends information to the OSD about changes in cell signal strength.

To use the application you will need a server. You must build an IPSEC tunnel with Rpi and you will also need the project https://github.com/bluenviron/mavp2p to receive telemetry.