package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"reflect"
	"time"
)

// TPV represents a gpsd TPV (time-position-velocity) class (https://gpsd.io/gpsd_json.html#_tpv)
type TPV struct {
	Device      string  `json:"device" description:"Name of the originating device"`
	Mode        float64 `json:"mode" description:"NMEA mode: 0=unknown, 1=no fix, 2=2D, 3=3D"`
	Status      float64 `json:"status" description:"GPS fix status: 0=Unknown, 1=Normal, 2=DGPS, 3=RTK Fixed, 4=RTK Floating, 5=DR, 6=GNSSDR, 7=Time (surveyed), 8=Simulated, 9=P(Y)"`
	Time        string  `json:"time" description:"Time/date stamp in ISO8601 format, UTC. May have a fractional part of up to .001sec precision. May be absent if the mode is not 2D or 3D. May be present, but invalid, if there is no fix. Verify 3 consecutive 3D fixes before believing it is UTC. Even then it may be off by several seconds until the current leap seconds is known."`
	AltHAE      float64 `json:"altHAE" description:"Altitude, Height Above Ellipsoid, in meters. Probably WGS84."`
	AltMSL      float64 `json:"altMSL" description:"MSL Altitude in meters. The geoid used is rarely specified and is often inaccurate. See the comments below on geoidSep. altMSL is altHAE minus geoidSep."`
	Climb       float64 `json:"climb" description:"Climb (positive) or sink (negative) rate, meters per second."`
	Datum       string  `json:"datum" description:"Current datum. Hopefully WGS84."`
	Depth       float64 `json:"depth" description:"Depth in meters. Probably depth below the keel"`
	DGPSAge     float64 `json:"dgpsAge" description:"Age of DGPS data in seconds"`
	DGPSStation float64 `json:"dgpsSta" description:"Station of DGPS data"`
	EPC         float64 `json:"epc" description:"Estimated climb error in meters per second. Certainty unknown."`
	EPD         float64 `json:"epd" description:"Estimated track (direction) error in degrees. Certainty unknown."`
	EPH         float64 `json:"eph" description:"Estimated horizontal Position (2D) Error in meters. Also known as Estimated Position Error (epe). Certainty unknown."`
	EPS         float64 `json:"eps" description:"Estimated speed error in meters per second. Certainty unknown."`
	EPT         float64 `json:"ept" description:"Estimated time stamp error in seconds. Certainty unknown."`
	EPX         float64 `json:"epx" description:"Longitude error estimate in meters. Certainty unknown."`
	EPY         float64 `json:"epy" description:"Latitude error estimate in meters. Certainty unknown."`
	EPV         float64 `json:"epv" description:"Estimated vertical error in meters. Certainty unknown."`
	GeoidSep    float64 `json:"geoidSep" description:"Geoid separation is the difference between the WGS84 reference ellipsoid and the geoid (Mean Sea Level) in meters. Almost no GNSS receiver specifies how they compute their geoid.gpsd interpolates the geoid from a 5x5 degree table of EGM2008 values when the receiver does not supply a geoid separation.The gpsd computed geoidSep is usually within one meter of the \"true\" value, but can be off as much as 12 meters."`
	Lat         float64 `json:"lat" description:"Latitude in degrees: +/- signifies North/South."`
	LeapSeconds float64 `json:"leapseconds" description:"Current leap seconds."`
	Lon         float64 `json:"lon" description:"Longitude in degrees: +/- signifies East/West."`
	Track       float64 `json:"track" description:"Course over ground, degrees from true north."`
	MagTrack    float64 `json:"magtrack" description:"Course over ground, degrees magnetic."`
	MagVar      float64 `json:"magvar" description:"Magnetic variation, degrees.Also known as the magnetic declination (the direction of the horizontal component of the magnetic field measured clockwise from north) in degrees, Positive is West variation.Negative is East variation."`
	Speed       float64 `json:"speed" description:"Speed over ground, meters per second."`
	ECEFX       float64 `json:"ecefx" description:"ECEF X position in meters."`
	ECEFY       float64 `json:"ecefy" description:"ECEF Y position in meters."`
	ECEFZ       float64 `json:"ecefz" description:"ECEF Z position in meters."`
	ECEFPAcc    float64 `json:"ecefpAcc" description:"ECEF position error in meters.Certainty unknown."`
	ECEFVX      float64 `json:"ecefvx" description:"ECEF X velocity in meters per second."`
	ECEFVY      float64 `json:"ecefvy" description:"ECEF Y velocity in meters per second."`
	ECEFVZ      float64 `json:"ecefvz" description:"ECEF Z velocity in meters per second."`
	ECEFVAcc    float64 `json:"ecefvAcc" description:"ECEF velocity error in meters per second. Certainty unknown."`
	Sep         float64 `json:"sep" description:"Estimated Spherical (3D) Position Error in meters.Guessed to be 95% confidence, but many GNSS receivers do not specify, so certainty unknown."`
	RelD        float64 `json:"relD" description:"Down component of relative position vector in meters."`
	RelE        float64 `json:"relE" description:"East component of relative position vector in meters."`
	RelN        float64 `json:"relN" description:"North component of relative position vector in meters."`
	VelD        float64 `json:"velD" description:"Down velocity component in meters."`
	VelE        float64 `json:"velE" description:"East velocity component in meters."`
	VelN        float64 `json:"velN" description:"North velocity component in meters."`
	WAngleM     float64 `json:"wanglem" description:"Wind angle magnetic in degrees."`
	WAngleR     float64 `json:"wangler" description:"Wind angle relative in degrees."`
	WAngleT     float64 `json:"wanglet" description:"Wind angle true in degrees."`
	WSpeedR     float64 `json:"wspeedr" description:"Wind speed relative in meters per second."`
	WSpeedT     float64 `json:"wspeedt" description:"Wind speed true in meters per second."`
}

// SKY represents a gpsd SKY (satellite position sky view) class (https://gpsd.io/gpsd_json.html#_sky)
type SKY struct {
	Device     string      `json:"device" description:"Name of originating device"`
	NSat       float64     `json:"nSat" description:"Number of satellite objects in \"satellites\" array."`
	GDOP       float64     `json:"gdop" description:"Geometric (hyperspherical) dilution of precision, a combination of PDOP and TDOP. A dimensionless factor which should be multiplied by a base UERE to get an error estimate."`
	HDOP       float64     `json:"hdop" description:"Horizontal dilution of precision, a dimensionless factor which should be multiplied by a base UERE to get a circular error estimate."`
	PDOP       float64     `json:"pdop" description:"Position (spherical/3D) dilution of precision, a dimensionless factor which should be multiplied by a base UERE to get an error estimate."`
	PRRes      float64     `json:"prRes" description:"Pseudorange residue in meters"`
	Qual       float64     `json:"qual" description:"Quality Indicator: 0 = no signal, 1 = searching signal, 2 = signal acquired, 3 = signal detected but unusable, 4 = code locked and time synchronized, 5, 6, 7 = code and carrier locked and time synchronized"`
	Satellites []Satellite `json:"satellites" description:"List of satellite objects in skyview"`
	TDOP       float64     `json:"tdop" description:"Time dilution of precision, a dimensionless factor which should be multiplied by a base UERE to get an error estimate."`
	Time       string      `json:"time" description:"Time/date stamp in ISO8601 format, UTC. May have a fractional part of up to .001sec precision."`
	USat       float64     `json:"uSat" description:"Number of satellites used in navigation solution."`
	VDOP       float64     `json:"vdop" description:"Vertical (altitude) dilution of precision, a dimensionless factor which should be multiplied by a base UERE to get an error estimate."`
	XDOP       float64     `json:"xdop" description:"Longitudinal dilution of precision, a dimensionless factor which should be multiplied by a base UERE to get an error estimate."`
	YDOP       float64     `json:"ydop" description:"Latitudinal dilution of precision, a dimensionless factor which should be multiplied by a base UERE to get an error estimate."`
}

// Satellite represents a gpsd Satellite (satellite object) class (https://gpsd.io/gpsd_json.html#_satellite)
type Satellite struct {
	PRN       float64 `json:"PRN" description:"PRN ID of the satellite. 1-63 are GNSS satellites, 64-96 are GLONASS satellites, 100-164 are SBAS satellites"`
	Azimuth   float64 `json:"az" description:"Azimuth, degrees from true north."`
	Elevation float64 `json:"el" description:"Elevation in degrees."`
	SNR       float64 `json:"ss" description:"Signal to Noise ratio in dBHz."`
	Used      bool    `json:"used"  description:"Used in current solution? (SBAS/WAAS/EGNOS satellites may be flagged used if the solution has corrections from them, but not all drivers make this information available.)"`
	GNSSID    float64 `json:"gnssid" description:"The GNSS ID, as defined by u-blox, not NMEA. 0=GPS, 2=Galileo, 3=Beidou, 5=QZSS, 6-GLONASS."`
	SVID      float64 `json:"svid" description:"The satellite ID within its constellation. As defined by u-blox, not NMEA)."`
	SigID     float64 `json:"sigid" description:"The signal ID of this signal. As defined by u-blox, not NMEA. See u-blox doc for details."`
	FreqID    float64 `json:"freqid" description:"For GLONASS satellites only: the frequency ID of the signal. As defined by u-blox, range 0 to 13. The freqid is the frequency slot plus 7."`
	Health    float64 `json:"health" description:"The health of this satellite. 0 is unknown, 1 is OK, and 2 is unhealthy."`
}

// GST represents a gpsd GST (pseudorange noise report) class (https://gpsd.io/gpsd_json.html#_gst)
type GST struct {
	Device string  `json:"device" description:"Name of originating device"`
	Time   string  `json:"time" description:"Time/date stamp in ISO8601 format, UTC. May have a fractional part of up to .001sec precision."`
	RMS    float64 `json:"rms" description:"Value of the standard deviation of the range inputs to the navigation process (range inputs include pseudoranges and DGPS corrections)."`
	Major  float64 `json:"major" description:"Standard deviation of semi-major axis of error ellipse, in meters."`
	Minor  float64 `json:"minor" description:"Standard deviation of semi-minor axis of error ellipse, in meters."`
	Orient float64 `json:"orient" description:"Orientation of semi-major axis of error ellipse, in degrees from true north."`
	Lat    float64 `json:"lat" description:"Standard deviation of latitude error, in meters."`
	Lon    float64 `json:"lon" description:"Standard deviation of longitude error, in meters."`
	Alt    float64 `json:"alt" description:"Standard deviation of altitude error, in meters."`
}

// TOFF represents a gpsd TOFF (time offset) class (https://gpsd.io/gpsd_json.html#_toff)
type TOFF struct {
	Device    string  `json:"device" description:"Name of the originating device"`
	RealSec   float64 `json:"real_sec" description:"seconds from the GPS clock"`
	RealNsec  float64 `json:"real_nsec" description:"nanoseconds from the GPS clock"`
	ClockSec  float64 `json:"clock_sec" description:"seconds from the system clock"`
	ClockNsec float64 `json:"clock_nsec" description:"nanoseconds from the system clock"`
}

// PPS represents a gpsd PPS (pulse per second) class (https://gpsd.io/gpsd_json.html#_pps)
type PPS struct {
	Device    string  `json:"device" description:"Name of the originating device"`
	RealSec   float64 `json:"real_sec" description:"seconds from the PPS source"`
	RealNsec  float64 `json:"real_nsec" description:"nanoseconds from the PPS source"`
	ClockSec  float64 `json:"clock_sec" description:"seconds from the system clock"`
	ClockNsec float64 `json:"clock_nsec" description:"nanoseconds from the system clock"`
	Precision float64 `json:"precision" description:"NTP style estimate of PPS precision"`
	SHM       string  `json:"shm" description:"shm key of this PPS"`
	QErr      float64 `json:"qErr" description:"Quantization error of the PPS, in picoseconds. Sometimes called the \"sawtooth\" error."`
}

// OSC represents a gpsd OSC (oscillator) class (https://gpsd.io/gpsd_json.html#_osc)
type OSC struct {
	Device      string  `json:"device" description:"Name of the originating device."`
	Running     bool    `json:"running" description:"If true, the oscillator is currently running. Oscillators may require warm-up time at the start of the day."`
	Reference   bool    `json:"reference" description:"If true, the oscillator is receiving a GPS PPS signal."`
	Disciplined bool    `json:"disciplined" description:"If true, the GPS PPS signal is sufficiently stable and is being used to discipline the local oscillator."`
	Delta       float64 `json:"delta" description:"The time difference (in nanoseconds) between the GPS-disciplined oscillator PPS output pulse and the most recent GPS PPS input pulse."`
}

func updateSatellite(sat *Satellite) {
	v := reflect.ValueOf(sat)
	for v.Kind() == reflect.Ptr { // Dereference pointer types
		v = v.Elem()
	}
	vType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		key := "gpsd_sat_" + vType.Field(i).Tag.Get("json")
		if key == "gpsd_sat_PRN" {
			continue
		}
		log.Tracef("%s = %+v\n", key, field.Interface())

		// Create the metrics if they don't exist
		switch field.Type().Kind() {
		case reflect.Bool, reflect.Float64:
			log.Tracef("Creating gaugevec metric %s", key)
			if _, exists := dynMetricGaugeVecs[key]; !exists {
				dynMetricGaugeVecs[key] = promauto.NewGaugeVec(prometheus.GaugeOpts{
					Name: key,
					Help: vType.Field(i).Tag.Get("description"),
				}, []string{"prn"})
			}
		default:
			log.Fatalf("Unsupported type %s for %s", field.Type().Kind(), key)
		}

		prnStr := fmt.Sprintf("%d", int(sat.PRN))

		// Update the metrics
		switch field.Type().Kind() {
		case reflect.Bool:
			if field.Interface().(bool) {
				log.Tracef("Setting %s to 1\n", key)
				dynMetricGaugeVecs[key].With(map[string]string{"prn": prnStr}).Set(1)
			} else {
				log.Tracef("Setting %s to 0\n", key)
				dynMetricGaugeVecs[key].With(map[string]string{"prn": prnStr}).Set(0)
			}
		case reflect.Float64:
			log.Tracef("Setting %s to %f\n", key, field.Interface().(float64))
			dynMetricGaugeVecs[key].With(map[string]string{"prn": prnStr}).Set(field.Interface().(float64))
		}
	}
}

func updateMetrics(t any, namespace string) {
	v := reflect.ValueOf(t)
	for v.Kind() == reflect.Ptr { // Dereference pointer types
		v = v.Elem()
	}
	vType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		key := fmt.Sprintf("gpsd_%s_%s", namespace, vType.Field(i).Tag.Get("json"))
		log.Tracef("%s = %+v\n", key, field.Interface())

		if field.Type().Kind() == reflect.String && vType.Field(i).Tag.Get("json") != "time" {
			continue
		}

		// Create the metrics if they don't exist
		switch field.Type().Kind() {
		case reflect.Bool, reflect.Float64, reflect.String:
			log.Tracef("Creating gauge metric %s", key)
			if _, exists := dynMetricGauges[key]; !exists {
				dynMetricGauges[key] = promauto.NewGauge(prometheus.GaugeOpts{
					Name: key,
					Help: vType.Field(i).Tag.Get("description"),
				})
			}
		case reflect.Slice:
			if key != "gpsd_sky_satellites" {
				log.Fatalf("Found slice that isn't a satellite slice: %s", key)
			}

			// Handle satellite slice
			for j := 0; j < field.Len(); j++ {
				satellite := field.Index(j).Interface().(Satellite)
				updateSatellite(&satellite)
			}
		default:
			log.Fatalf("Unsupported type %s for %s", field.Type().Kind(), key)
		}

		// Update the metrics
		switch field.Type().Kind() {
		case reflect.Bool:
			if field.Interface().(bool) {
				log.Tracef("Setting %s to 1\n", key)
				dynMetricGauges[key].Set(1)
			} else {
				log.Tracef("Setting %s to 0\n", key)
				dynMetricGauges[key].Set(0)
			}
		case reflect.String:
			timeStr := field.Interface().(string)
			if timeStr != "" {
				timestamp, err := time.Parse(time.RFC3339Nano, timeStr)
				if err != nil {
					log.Fatalf("Failed to parse time %s: %s", timeStr, err)
				}
				dynMetricGauges[key].Set(float64(timestamp.UnixNano()))
			}
		case reflect.Float64:
			log.Tracef("Setting %s to %f\n", key, field.Interface().(float64))
			dynMetricGauges[key].Set(field.Interface().(float64))
		}
	}
}

func processLine(line string) {
	if len(line) < 16 {
		return
	}
	var f interface{}
	if err := json.Unmarshal([]byte(line), &f); err != nil {
		log.Fatal(err)
	}

	m := f.(map[string]interface{})
	cl := m["class"]
	switch cl {
	case "VERSION":
		metricVersion.With(
			map[string]string{
				"version": fmt.Sprintf("GPSD v%s", m["release"].(string)),
			},
		).Set(1)
	case "POLL":
		for pollClass := range m {
			switch pollClass {
			case "sky":
				var skyFrame struct {
					Sky []SKY `json:"sky"`
				}
				if err := json.Unmarshal([]byte(line), &skyFrame); err != nil {
					log.Warnf("Error unmarshalling SKY: %v", err)
				}
				log.Tracef("SKY: %+v", skyFrame.Sky)
				for _, sky := range skyFrame.Sky {
					updateMetrics(sky, "sky")
				}
			case "tpv":
				var tpvFrame struct {
					TPV []TPV `json:"tpv"`
				}
				if err := json.Unmarshal([]byte(line), &tpvFrame); err != nil {
					log.Warnf("Error unmarshalling TPV: %v", err)
				}
				log.Tracef("TPV: %+v", tpvFrame.TPV)
				for _, tpv := range tpvFrame.TPV {
					updateMetrics(tpv, "tpv")
				}
			case "gst":
				var gstFrame struct {
					GST []GST `json:"gst"`
				}
				if err := json.Unmarshal([]byte(line), &gstFrame); err != nil {
					log.Warnf("Error unmarshalling GST: %v", err)
				}
				log.Tracef("GST: %+v", gstFrame.GST)
				for _, gst := range gstFrame.GST {
					updateMetrics(gst, "gst")
				}
			case "pps":
				var ppsFrame struct {
					PPS []PPS `json:"pps"`
				}
				if err := json.Unmarshal([]byte(line), &ppsFrame); err != nil {
					log.Warnf("Error unmarshalling PPS: %v", err)
				}
				log.Tracef("PPS: %+v", ppsFrame.PPS)
				for _, pps := range ppsFrame.PPS {
					updateMetrics(pps, "pps")
				}
			case "toff":
				var toffFrame struct {
					Toff []TOFF `json:"toff"`
				}
				if err := json.Unmarshal([]byte(line), &toffFrame); err != nil {
					log.Warnf("Error unmarshalling TOFF: %v", err)
				}
				log.Tracef("TOFF: %+v", toffFrame.Toff)
				for _, toff := range toffFrame.Toff {
					updateMetrics(toff, "toff")
				}
			case "osc":
				var oscFrame struct {
					OSC []OSC `json:"osc"`
				}
				if err := json.Unmarshal([]byte(line), &oscFrame); err != nil {
					log.Warnf("Error unmarshalling OSC: %v", err)
				}
				log.Tracef("OSC: %+v", oscFrame.OSC)
				for _, osc := range oscFrame.OSC {
					updateMetrics(osc, "osc")
				}
			case "class", "active", "time":
				// Ignore
			default:
				log.Printf("Unknown poll type: %s in line %s", pollClass, line)
			}
		}
	}
}
