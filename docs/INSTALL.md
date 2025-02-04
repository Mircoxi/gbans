# Installation Guide

Basic installation overview of the gbans server and sourcemod plugin.

## System Considerations

Gbans is lightweight and can handle a small to moderately sized community with a dual-core CPU and 4GB of memory.

Special considerations need to be made when using extended functionality: 

- STV demo recording can fill space up quickly. STVs are stored in the database directly, and typically removed after 
  two weeks. Allow 15GB or so per TF2 server instance for demo recordings. Space can be reduced by enabling TOAST compression in
  Postgres. 
- IP2Location is memory intensive when updating the dataset, requiring 10 to 12GB of memory. The process can be
  sped up by using NVMe storage for the database. 

Larger communities will inherently require more resources.

## Sourcemod Plugins

The following extensions must be installed for gbans to work, see their documentation for up to date installation
instructions:

- [System2](https://github.com/dordnung/System2) Provides HTTP(S) client functionality
- [sm-json](https://github.com/clugg/sm-json) `Required for development only` Provides JSON encoding/decoding.
- [Connect](https://github.com/asherkin/connect) Provides `OnClientPreConnectEx`

## gbans Server

### Compile from source
Precompiled binaries will be provided once the project is in a more stable state.

- [make](https://www.gnu.org/software/make/) Not strictly required but provides predefined build commands
- [golang 1.16+](https://golang.org/) gbans is written in go. Version >=1.16 is *REQUIRED* due to using iofs embed
  features.
- [PostgreSQL](https://www.postgresql.org/) is used as the data store. Version 12 is the only version currently tested
  against. However i believe anything 10 and up should work. Please let me know if this is not the case.
    - [PostGIS](https://postgis.net/) extension is also used for some GIS functionality.
- [NodeJS 14+](https://nodejs.org/en/) To build frontend
    - [yarn](https://yarnpkg.com/) JS package manager

Basic steps to build the binary packages:

    1. git clone git@github.com:leighmacdonald/gbans.git && cd gbans
    2. make

You should now have a binary located at `./build/$platform/gbans`

### Docker 

```
sudo docker run -d --restart unless-stopped \
    -p 6006:6006 \
    --dns=1.1.1.1 \
    -v /home/ubuntu/gbans/gbans.yml:/app/gbans.yml:ro \
    --name gbans \
    ghcr.io/leighmacdonald/gbans:master
```

Substitute `master` for a specific tag if desired, and `/home/ubuntu/gbans/gbans.yml` with the location of your config.

This configuration will restart gbans unless explicitly stopped, and names the container for easy log access/stopping.

## Configuration

### Server

Copy the example configuration `gbans_example.yml` and name it `gbans.yml`. It should be in
the same directory as the binary. Configure it as desired.

#### Starting the server

To start the server just run `./gbans serve`. It should show output similar to the following if
successful.

```
➜  gbans git:(master) ✗ ./gbans serve
INFO[0000] Using config file: gbans.yaml 
INFO[0000] Starting gbans service                       
DEBU[0000] Ban sweeper routine started                  
INFO[0000] Bot is now running.  Press CTRL-C to exit.   
INFO[0000] Connected to session ws API                  
```

It's recommended to create a [systemd .service](https://freedesktop.org/software/systemd/man/systemd.service.html)
file so that it can start automatically. More info on configuring this will be available at a later
date.

### Sourcemod

Place the `sourcemod/plugins/gbans.smx` file into `tf/addons/sourcemod/plugins`. Then add the config as
described below.

This config file should be places in `tf/addons/sourcemod/configs/gbans.cfg`.

```
"gbans"
{
	// Remote gban server host
	"host"	"https://gbans.example.com"

	// Remote gban server port
	"port"	"443"

	// Unique server name for this server, the same as a "server-id"
	"server_name"	"example-1"

	// The authentication token used to retrieve a auth token
	"server_key"	"YOUR_TOKEN"
}
```

The server gbans server is running you should now be able to see the `[GB]` message logs in the
console. With a message like below on successful authentication with the server.

```
[GB] Using config file: addons/sourcemod/configs/gbans.cfg
[GB] Request to https://gbans.example.com/v1/auth finished with status code 200 in 0.01 seconds
[GB] Successfully authenticated with gbans server
```

### Discord

To use discord you need to [create a discord application](https://discord.com/developers/applications). You will need
the
following values from your application:

- Application ID (General -> Application ID)
- Token  (Bot -> Token)
- Client Secret (OAuth2 -> Client Secret)

You Will also need to fetch the following ids from your discord client. You will want to enable discord developer mode
to be able to easily acquire these role and channel ids through your own discord client.

- Your main server guild id.
- Logging Channels IDS
    - Public Log Channel
    - (Private) Moderation Channel
    - (Private) Bot Logs
    - (Private) Report Logs
- Moderator Roles Ids

You must also set an oauth2 redirect (Oauth2 -> Redirects -> Add) to point to your own server.

    https://example.com/login/discord

Example configuration for discord

    discord:
      # Enable optional discord integration
      enabled: true
      app_id: 814566730000000000
      app_secret: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
      guild_id: 875964612233801748
      # Your discord bot token
      # See: https://discord.com/developers/applications
      token: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
      mod_role_ids: [333333333333333333, 4444444444444444444]
      # People in these channels automatically have moderator privilege
      # To find these, Right click the channel -> copy id
      mod_channel_ids:
        - "111111111111111111"
      mod_log_channel_id: "111111111111111111"
      log_channel_id: "111111111111111111"
      public_log_channel_enable: true
      public_log_channel_id: "222222222222222222"
      report_log_channel_id: "111111111111111111"

## Reverse Proxy

### Caddy w/cloudflare

    example.com {
        reverse_proxy /* internal_host:6006
        encode gzip
        tls your@email.com {
            dns cloudfalre your_api_token
        }
    }

## Apache 2.4

Be sure to run `sudo a2enmod proxy_http ssl` first.
    
```
<IfModule mod_ssl.c>
<VirtualHost *:443>
        ServerName example.com

        ProxyPass / http://127.0.0.1:6006/
        ProxyPassReverse / http://127.0.0.1:6006/

        ServerAdmin your@email.com

        #Can be disabled if wanted
        ErrorLog ${APACHE_LOG_DIR}/error.log
        CustomLog ${APACHE_LOG_DIR}/access.log combined
        
SSLCertificateFile /etc/cloudflare/example.com.pem
SSLCertificateKeyFile /etc/cloudflare/example.com.key
</VirtualHost>
</IfModule>
```
If using Cloudflare to provide user location, you can use Origin Certificates to generate a long-lasting SSL certicate.
  

## IP2Location 

To install the GeoLite2 databases, create an account on [IP2location Lite](https://lite.ip2location.com). After
confirmation, you'll be given a download token for use in gbans.yaml.

If using Docker, open a terminal with `docker exec -it gbans /bin/sh`, or if using a compiled binary, navigate to the
folder. 

Run `./gbans net update` to start the process. This will require around 12GB of memory (or a suitably large swapfile), 
and does *not* need to be run on the host - a more powerful machine can run it, as long as the config is 
mirrored and database access works.

The process will take up to 30 minutes, depending on hardware, and will add around 2GB to the database when all's said
and done.

## Enabling User Location

The Servers page lets users sort by range. Gbans does not use the locations API to get data from the browser. 
Instead, you're required to use Cloudflare to get the location. Gbans must be proxied through Cloudflare to
accomplish this, and setting that up is out of scope of this doc. 

Once the domain is set up, go to the domain settings, the `Rules` dropdown, `Transform Rules`, and then the 
`Managed Transforms` tab. Enable `Add visitor location headers`, and wait around 5 minutes for it to take effect. 
You should then be able to see your location (more or less) on the servers page.