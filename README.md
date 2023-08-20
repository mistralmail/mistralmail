![](.github/workflows/mistralmail128.png)

# MistralMail

MistralMail will be a production-ready, and easy to setup mail server. It consists of an SMTP server (both MSA and MTA) and an IMAP server all bundled in one executable (or just in one Docker image) with auto-generated TLS certificates.

⚠️ **WIP: MistralMail is far from being production-ready!** ⚠️



## Usage

### Setting up DNS records

MistralMail will not be able to generate TLS certificates without a correct DNS configuration. And of course you also won't be able to receive any emails. (But if you just want to configure it locally you can set `TLS_DISABLE` to `true` and skip this section.)

You need the following DNS records:

- A record pointing to your MistralMail server ip.

- MX record point to your MistralMail server ip.

- SPF record pointing to your SMTP relay provider.

### Running the MistralMail server

First you need to copy `.env.sample` to `.env` and configure all the needed environment variables.

Make sure that ports 80 and 443 are opened for the automatic TLS certificate generation via Let's Encrypt.

Then you can run the Go main manually or with Docker.

**Go:**

```bash
source .env
go run cmd/mistralmail/*.go
```

**Docker:**

Everything needed is put into the `docker-compose.yml` file.
If you don't want to build the image yourself you can use the prebuilt one present at `denbeke/mistralmail`.

```bash
docker-compose up mistralmail
```

Now you can create a user with the MistralMail CLI.

MistralMail exposes the following ports:

- 25 for all incoming SMTP emails (MTA)

- 587 for all outing SMTP emails (MSA)

- 143 for IMAP

- 443 & 80 for Let's Encrypt

### Using the MistralMail command line interface

You can use the MistralMail command line interface with Go or with Docker:

```bash
go run cmd/mistralmail-cli/*.go
```

or

```bash
docker-compose run mistralmail mistralmail-cli
```

Currently the CLI contains the following commands:

- `create-user`: to create a new user.

- `reset-password` to reset the password of a user.

### Configuring your mail client

**IMAP:**

- **Server address:** your domain name

- **Username:** your email address

- **Port:** 143

- **Security:** STARTTLS

- **Authentication:** password

**SMTP:**

- **Server address:** your domain name

- **Username:** your email address

- **Port:** 587

- **Security:** STARTTLS

- **Authentication:** password



Now you're all good to go!



## Development

We use `go work` for updating files across multiple repo's:

```bash
go work init
go work use smtp
go work use imap-backend
```



## Current state of MistralMail

### SMTP server

The SMTP server is completely custom written and can be found here: [mistralmail/smtp](https://github.com/mistralmail/smtp). It was written quite a while ago but it seems robust enough for now.

For outgoing emails we currently only support using an external relay like Mailgun or Sendgrid since we don't want to put too much time into debugging an MSA.

### IMAP

For IMAP we wrote a SQL backend behind [go-imap](https://github.com/emersion/go-imap). It supports MySQL, Postgres and Sqlite. (Currently only Sqlite has actually been tested.)

This backend is very experimental and surely contains a lot of bug. The backend is also implemented in a very non-performant way. So don't expect that MistralMail will be able to handle large inboxes at its current state.

We dump the complete emails in the database at this moment. In the future we would like to add support for object storage for the actual mail bodies. But that's nothing for the near future.

## Webmail

Currently there are no concrete plans to implement a webmail. But wouldn't it be nice to have it someday?

### Web management

Instead of configuring everything via a CLI, we'd like to add an admin dashboard on which you can easily configure everything. Checking DNS records, checking state of the server, metrics, managing users, ...

### SPAM

Another feature we are also not working on currently is anti-spam. Only SPF is checked at the moment. But nothing else.



## Acknowledgements

* [Testify](https://github.com/stretchr/testify)
* [GoConvey](https://github.com/smartystreets/goconvey)
* [go-imap](https://github.com/emersion/go-imap)



## Authors

[Mathias Beke](https://denbeke.be)
Timo Truyts
