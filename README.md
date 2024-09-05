![image](https://github.com/user-attachments/assets/e19ac0b5-45c0-4144-930d-5d4a7fbeb5b6)

# 📻 Radio

A web player for (icecast) radio streams with integrated chat.

### ⚙️ Configuration

1. Configure the stations to broadcast in a ini file (See [example](./stations.ini)).
2. Set the path to the stations file and the address to listen on via
   environment variables (`RADIO_ADDRESS` and `RADIO_STATIONFILE`).
3. Make sure your reverse-proxy proxies websockets.

Optionl: Set `RADIO_PROXY_STATIONS` to true to proxy all music streams through
the server. This causes a higher load on the server but can work around CORS
problems. If you have no problems, leave it set to the default (false).

### 🎵 Stations

If you have an icecast stream and would like to get a station on
[0cx.radio.de](https://radio.0cx.de) feel free to reach out and you may get
featured.
