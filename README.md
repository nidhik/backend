# backend
Gin + Mongo backend with Facebook &amp; email authentication. Replaces much of the functionality of the defunct Parse Baas

Sample Env file:

PATH=$GOROOT/bin:/usr/local/bin:$PATH

LITEIDE_GDB=/usr/local/bin/gdb
LITEIDE_MAKE=make
LITEIDE_TERM=/usr/bin/open
LITEIDE_TERMARGS=-a Terminal
LITEIDE_EXEC=/usr/X11R6/bin/xterm
LITEIDE_EXECOPT=-e

TMPL_DIR=templates
PROCESS_TYPE=web
#PROCESS_TYPE=worker
DB_CONNECTION_URL=mongodb://<dbuser>:<dbpass>@<db_url>/<db_name>
JWT_SECRET=<yoursecret>
ENV_TYPE=sandbox
CLIENT_KEY=<devclientkey>
FB_APP_ACCESS_TOKEN=<fbappid>|<fbappsecret>
MY_FB_TOKEN=<fb-oauthtoken>
MY_FB_ID=<fb-id>
ALLOWED_ORIGIN=*
