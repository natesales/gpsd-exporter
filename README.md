# Prometheus exporter for [gpsd](https://gpsd.io)

[![Release](https://img.shields.io/github/v/release/natesales/gpsd-exporter?style=for-the-badge)](https://github.com/natesales/gpsd-exporter/releases)
[![Go Report](https://goreportcard.com/badge/github.com/natesales/gpsd-exporter?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/gpsd-exporter)
[![License](https://img.shields.io/github/license/natesales/gpsd-exporter?style=for-the-badge)](https://raw.githubusercontent.com/natesales/gpsd-exporter/main/LICENSE)

`gpsd-exporter` polls `gpsd` over its TCP JSON interface and exports the data to Prometheus.

### Supported gpsd classes

- Time position value ([TPV](https://gpsd.io/gpsd_json.html#_tpv))
- Sky view ([SKY](https://gpsd.io/gpsd_json.html#_sky))
- Satellite ([Satellite](https://gpsd.io/gpsd_json.html#_satellite))
- Pseudorange noise report ([GST](https://gpsd.io/gpsd_json.html#_gst))
- Time offset ([TOFF](https://gpsd.io/gpsd_json.html#_toff))
- Pulse per second ([PPS](https://gpsd.io/gpsd_json.html#_pps))
- Oscillator ([OSC](https://gpsd.io/gpsd_json.html#_osc))
- gpsd Version ([VERSION](https://gpsd.io/gpsd_json.html#_version))

See [gpsd's protocol responses](https://gpsd.io/gpsd_json.html#_core_protocol_responses) for more information.
