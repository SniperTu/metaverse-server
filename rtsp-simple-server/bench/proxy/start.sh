#!/bin/sh -e

PROXY_COUNT=50
PROXY_PROTOCOL=tcp

#####################################################
# source

CONF=""
CONF="${CONF}rtmp: no\n"
CONF="${CONF}hls: no\n"
CONF="${CONF}webrtc: no\n"
CONF="${CONF}srt: no\n"
CONF="${CONF}rtspAddress: :8555\n"
CONF="${CONF}rtpAddress: :8002\n"
CONF="${CONF}rtcpAddress: :8003\n"
CONF="${CONF}paths:\n"
CONF="${CONF}  all_others:\n"
echo -e "$CONF" > /source.conf

/mediamtx /source.conf &

sleep 1

ffmpeg -hide_banner -loglevel error \
-re -stream_loop -1 -i /video.mkv -c copy -f rtsp rtsp://localhost:8555/source &

sleep 1

#####################################################
# proxy

CONF=""
CONF="${CONF}rtmp: no\n"
CONF="${CONF}hls: no\n"
CONF="${CONF}webrtc: no\n"
CONF="${CONF}srt: no\n"
CONF="${CONF}pprof: yes\n"
CONF="${CONF}paths:\n"
for i in $(seq 1 $PROXY_COUNT); do
    CONF="${CONF}  proxy$i:\n"
    CONF="${CONF}    source: rtsp://localhost:8555/source\n"
    CONF="${CONF}    rtspTransport: $PROXY_PROTOCOL\n"
done
echo -e "$CONF" > /proxy.conf

/mediamtx /proxy.conf &

sleep 5

go tool pprof -text http://localhost:9999/debug/pprof/profile?seconds=15
