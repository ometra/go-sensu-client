Configuration Snippets
----------------------

The Sensu client package creates a `/etc/sensu/config.json.dist` file that
contains example config stanzas required for a minimal Sensu client 
installation.

However, you may also break this monolithic config file up into smaller
pieces which often helps with config management systems such as Puppet or Chef.

Place JSON snippets in the `/etc/sensu/conf.d` directory. Files must have
a `.json` suffix.

Examples:
`/etc/sensu/conf.d/client.json`:

	{
		"client": {
			"name": "localhost",
			"address": "192.168.1.1",
			"subscriptions": [
				"test"
			]
		}
	}

`/etc/sensu/conf.d/rabbitmq.json`:

	{
		"rabbitmq": {
			"host": "localhost",
			"port": 5672,
			"user": "guest",
			"password": "guest",
			"vhost": "/sensu"
		}
	}