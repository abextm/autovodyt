set -e

cd "$(dirname "$0")"

source ./config.sh

go build -o ytupload .
./ytupload test

while ~/.local/bin/streamlink \
	--twitch-oauth-token "$TWITCH_OAUTH_TOKEN" \
	--twitch-disable-hosting \
	--player ./ytupload \
	--player-no-close \
	--verbose-player \
	--retry-streams 30 \
	"$STREAM_LINK" "best,high,1080p,720p"
	do
	true
done