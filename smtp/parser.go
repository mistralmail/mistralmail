package smtp

import "bufio"
import "strings"
import "errors"

type parser struct {
}

func (p *parser) ParseCommand(br *bufio.Reader) (command Cmd, err error) {
	/*
		RFC 5321 2.3.8

		Lines consist of zero or more data characters terminated by the
		sequence ASCII character "CR" (hex value 0D) followed immediately by
		ASCII character "LF" (hex value 0A).  This termination sequence is
		denoted as <CRLF> in this document.  Conforming implementations MUST
		NOT recognize or generate any other character or character sequence
		as a line terminator.  Limits MAY be imposed on line lengths by
		servers (see Section 4).
	*/

	line, _ := br.ReadString('\n')

	for line == "" {
		line, _ = br.ReadString('\n')
	}

	var address *MailAddress
	verb, args, err := parseLine(line)
	if err != nil {
		return nil, err
	}
	//conn.write(500, err.Error())
	//conn.c.Close()
	_ = args

	switch verb {

	case "HELO":
		{
			command = HeloCmd{}
		}

	case "EHLO":
		{
			command = EhloCmd{}
		}

	case "MAIL":
		{
			address, err = parseFROM(args)
			command = MailCmd{From: address}
		}

	case "RCPT":
		{
			address, err = parseTO(args)
			command = RcptCmd{To: address}
		}

	case "DATA":
		{
			messageContent := []byte{}

			for {
				data, _ := br.ReadString('\n')

				if data == ".\r\n" || data == ".\r" || data == ".\n" {
					// the character sequence "<CRLF>.<CRLF>" ends the mail text
					break
				}
				if data[0:2] == ".." {
					/*
						RFC 5321 4.5.2

						Without some provision for data transparency, the character sequence
						"<CRLF>.<CRLF>" ends the mail text and cannot be sent by the user.
						In general, users are not aware of such "forbidden" sequences.  To
						allow all user composed text to be transmitted transparently, the
						following procedures are used:
						o  Before sending a line of mail text, the SMTP client checks the
						   first character of the line.  If it is a period, one additional
						   period is inserted at the beginning of the line.
						o  When a line of mail text is received by the SMTP server, it checks
						   the line.  If the line is composed of a single period, it is
						   treated as the end of mail indicator.  If the first character is a
						   period and there are other characters on the line, the first
						   character is deleted.
					*/
					data = data[1:len(data)]
				}

				// merge mail data
				messageContent = append(messageContent, []byte(data)...)

				// continue to get data
				continue

				// TODO break when there is no more content
				// TODO check for content too long
				/*
					RFC 5321:
					The maximum total length of a message content (including any message
					header section as well as the message body) MUST BE at least 64K
					octets.  Since the introduction of Internet Standards for multimedia
					mail (RFC 2045 [21]), message lengths on the Internet have grown
					dramatically, and message size restrictions should be avoided if at
					all possible.  SMTP server systems that must impose restrictions
					SHOULD implement the "SIZE" service extension of RFC 1870 [10], and
					SMTP client systems that will send large messages SHOULD utilize it
					when possible.

					552 Too much mail data
				*/
			}

			command = DataCmd{Data: messageContent}

		}

	case "RSET":
		{
			command = RsetCmd{}
		}

	case "SEND":
		{
			command = SendCmd{}
		}

	case "SOML":
		{
			command = SomlCmd{}
		}

	case "SAML":
		{
			command = SamlCmd{}
		}

	case "VRFY":
		{
			//conn.write(502, "Command not implemented")
			/*
					RFC 821
					SMTP provides as additional features, commands to verify a user
					name or expand a mailing list.  This is done with the VRFY and
					EXPN commands
					RFC 5321
					As discussed in Section 3.5, individual sites may want to disable
					either or both of VRFY or EXPN for security reasons (see below).  As
					a corollary to the above, implementations that permit this MUST NOT
					appear to have verified addresses that are not, in fact, verified.
					If a site disables these commands for security reasons, the SMTP
					server MUST return a 252 response, rather than a code that could be
					confused with successful or unsuccessful verification.
					Returning a 250 reply code with the address listed in the VRFY
					command after having checked it only for syntax violates this rule.
					Of course, an implementation that "supports" VRFY by always returning
					550 whether or not the address is valid is equally not in
					conformance.
				From what I have read, 502 is better than 252...
			*/
			command = VrfyCmd{}
		}

	case "EXPN":
		{
			command = ExpnCmd{}
		}

	case "NOOP":
		{
			command = NoopCmd{}
		}

	case "QUIT":
		{
			command = QuitCmd{}
		}

	default:
		{
			command = InvalidCmd{Cmd: line}
		}

	}

	return
}

// parseLine returns the verb of the line and a list of all comma separated arguments
func parseLine(line string) (verb string, args []string, err error) {

	/*
		RFC 5321
		4.5.3.1.4.  Command Line

		The maximum total length of a command line including the command word
		and the <CRLF> is 512 octets.  SMTP extensions may be used to
		increase this limit.
	*/
	if len(line) > 512 {
		return "", []string{}, errors.New("Line too long")
	}

	i := strings.Index(line, " ")
	if i == -1 {
		verb = strings.ToUpper(strings.TrimSpace(line))
		return
	}

	verb = strings.ToUpper(line[:i])
	args = strings.Split(strings.TrimSpace(line[i+1:len(line)]), " ")
	return
}

func parseFROM(args []string) (*MailAddress, error) {
	if len(args) < 1 {
		return nil, errors.New("No FROM given")
	}

	joined_args := strings.Join(args, " ")
	index := strings.Index(joined_args, ":")
	if index == -1 {
		return nil, errors.New("No FROM given (didn't find ':')")
	}
	address_str := joined_args[index+1 : len(joined_args)]

	address, err := ParseAddress(address_str)
	if err != nil {
		return nil, err
	} else {
		return &address, nil
	}

}

func parseTO(args []string) (*MailAddress, error) {
	if len(args) < 1 {
		return nil, errors.New("No TO given")
	}

	joined_args := strings.Join(args, " ")
	index := strings.Index(joined_args, ":")
	if index == -1 {
		return nil, errors.New("No TO given (didn't find ':')")
	}
	address_str := joined_args[index+1 : len(joined_args)]

	address, err := ParseAddress(address_str)
	if err != nil {
		return nil, err
	} else {
		return &address, nil
	}
}
