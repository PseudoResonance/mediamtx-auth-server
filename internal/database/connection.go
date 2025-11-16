package database

import (
	"fmt"
	"log"
	"net/http"

	ttlcache "github.com/jellydator/ttlcache/v3"
)

/*
Connection/disconnection notifications from MediaMTX
*/
type Connection struct {
	Id       string
	Protocol string
}

/*
Wrapper to hold full connection details for retrieval when disconnecting users
*/
type ConnectionRecord struct {
	Info  Connection
	Creds *Credentials
}

func (d *DatabaseManager) revoke(connData *CredentialData) {
	conns := connData.getAndClearConnections()
	for _, conn := range conns {
		ret, exist := d.connections.GetAndDelete(conn)
		if ret == nil || !exist {
			// Untracked connection (ex: localhost, failed auth)
			continue
		}
		d.closeConnection(ret.Value().Info)
	}
}

/*
Called by MediaMTX onConnect webhook
*/
func (d *DatabaseManager) Connect(conn Connection) {
	ret := d.connections.Get(conn.Id)
	if ret == nil {
		// Untracked connection (ex: localhost, failed auth)
		return
	}
	//TODO there is a race condition where this may run before the auth connection is created
	wrapper := ret.Value()
	wrapper.Info.Protocol = conn.Protocol
	d.connections.Set(conn.Id, wrapper, ttlcache.PreviousOrDefaultTTL)
}

/*
Called by MediaMTX onDisconnect webhook
*/
func (d *DatabaseManager) Disconnect(conn Connection) {
	ret, exist := d.connections.GetAndDelete(conn.Id)
	if ret == nil || !exist {
		// Untracked connection (ex: localhost, failed auth)
		return
	}

	if creds := ret.Value().Creds; creds != nil {
		if credData := d.cache.Get(*creds); credData != nil {
			credData.Value().removeConnection(conn.Id)
		}
	}
}

/*
Tell MediaMTX to forcefully close a connection
*/
func (d *DatabaseManager) closeConnection(conn Connection) {
	switch conn.Protocol {
	case "hls":
		// Theoretically this shouldn't be used, but is here for just in case the onConnect webhook is missed for some reason
		fallthrough
	case "hlsMuxer":
		// Ignore as it's not persistent
		break

	case "rtmp":
		// Theoretically this shouldn't be used, but is here for just in case the onConnect webhook is missed for some reason
		d.postToUrl(fmt.Sprintf("%v/v3/rtmpconns/kick/%v", d.conf.MediaMtxUrlBase, conn.Id))
		d.postToUrl(fmt.Sprintf("%v/v3/rtmpsconns/kick/%v", d.conf.MediaMtxUrlBase, conn.Id))
	case "rtmpConn":
		d.postToUrl(fmt.Sprintf("%v/v3/rtmpconns/kick/%v", d.conf.MediaMtxUrlBase, conn.Id))
	case "rtmpsConn":
		d.postToUrl(fmt.Sprintf("%v/v3/rtmpsconns/kick/%v", d.conf.MediaMtxUrlBase, conn.Id))

	case "rtsp":
		// Theoretically this shouldn't be used, but is here for just in case the onConnect webhook is missed for some reason
		d.postToUrl(fmt.Sprintf("%v/v3/rtspsessions/kick/%v", d.conf.MediaMtxUrlBase, conn.Id))
		d.postToUrl(fmt.Sprintf("%v/v3/rtspssessions/kick/%v", d.conf.MediaMtxUrlBase, conn.Id))
	case "rtspSession":
		d.postToUrl(fmt.Sprintf("%v/v3/rtspsessions/kick/%v", d.conf.MediaMtxUrlBase, conn.Id))
	case "rtspsSession":
		d.postToUrl(fmt.Sprintf("%v/v3/rtspssessions/kick/%v", d.conf.MediaMtxUrlBase, conn.Id))

	case "srt":
		// Theoretically this shouldn't be used, but is here for just in case the onConnect webhook is missed for some reason
		fallthrough
	case "srtConn":
		d.postToUrl(fmt.Sprintf("%v/v3/srtconns/kick/%v", d.conf.MediaMtxUrlBase, conn.Id))

	case "webrtc":
		// Theoretically this shouldn't be used, but is here for just in case the onConnect webhook is missed for some reason
		fallthrough
	case "webRTCSession":
		d.postToUrl(fmt.Sprintf("%v/v3/webrtcsessions/kick/%v", d.conf.MediaMtxUrlBase, conn.Id))

	default:
		log.Printf("Unknown connection type %v", conn.Protocol)
	}
}

/*
Create a POST request with no body to the given URL
*/
func (d *DatabaseManager) postToUrl(url string) {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Printf("Error while closing connection\n%v\n", err)
	}
	d.httpClient.Do(req)
}

func (d *DatabaseManager) validateConnection(conn Connection) bool {
	switch conn.Protocol {
	// HLS shouldn't be in here at all as it's not persistent
	case "hls":
		fallthrough
	case "hlsMuxer":
		return false

	case "rtmp":
		// Theoretically this shouldn't be used, but is here for just in case the onConnect webhook is missed for some reason
		return d.getUrlValid(fmt.Sprintf("%v/v3/rtmpconns/get/%v", d.conf.MediaMtxUrlBase, conn.Id)) ||
			d.getUrlValid(fmt.Sprintf("%v/v3/rtmpsconns/get/%v", d.conf.MediaMtxUrlBase, conn.Id))
	case "rtmpConn":
		return d.getUrlValid(fmt.Sprintf("%v/v3/rtmpconns/get/%v", d.conf.MediaMtxUrlBase, conn.Id))
	case "rtmpsConn":
		return d.getUrlValid(fmt.Sprintf("%v/v3/rtmpsconns/get/%v", d.conf.MediaMtxUrlBase, conn.Id))

	case "rtsp":
		// Theoretically this shouldn't be used, but is here for just in case the onConnect webhook is missed for some reason
		return d.getUrlValid(fmt.Sprintf("%v/v3/rtspsessions/get/%v", d.conf.MediaMtxUrlBase, conn.Id)) ||
			d.getUrlValid(fmt.Sprintf("%v/v3/rtspssessions/get/%v", d.conf.MediaMtxUrlBase, conn.Id))
	case "rtspSession":
		return d.getUrlValid(fmt.Sprintf("%v/v3/rtspsessions/get/%v", d.conf.MediaMtxUrlBase, conn.Id))
	case "rtspsSession":
		return d.getUrlValid(fmt.Sprintf("%v/v3/rtspssessions/get/%v", d.conf.MediaMtxUrlBase, conn.Id))

	case "srt":
		// Theoretically this shouldn't be used, but is here for just in case the onConnect webhook is missed for some reason
		fallthrough
	case "srtConn":
		return d.getUrlValid(fmt.Sprintf("%v/v3/srtconns/get/%v", d.conf.MediaMtxUrlBase, conn.Id))

	case "webrtc":
		// Theoretically this shouldn't be used, but is here for just in case the onConnect webhook is missed for some reason
		fallthrough
	case "webRTCSession":
		return d.getUrlValid(fmt.Sprintf("%v/v3/webrtcsessions/get/%v", d.conf.MediaMtxUrlBase, conn.Id))

	default:
		log.Printf("Unknown connection type %v", conn.Protocol)
	}
	return false
}

/*
Create a GET request to the given URL and return if the code was OK
*/
func (d *DatabaseManager) getUrlValid(url string) bool {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error while checking if connection is valid\n%v\n", err)
	}
	res, err := d.httpClient.Do(req)
	if err != nil {
		log.Printf("Error while checking if connection is valid\n%v\n", err)
	}
	return res.StatusCode == http.StatusOK
}
