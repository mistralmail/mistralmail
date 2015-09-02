package smtp

import "bufio"
import "strings"
import "errors"

type parser struct {
}

func (p *parser) ParseCommand(br *bufio.Reader) (error, Cmd) {
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

	if line == "" {
		return nil, nil
	}

	verb, args, err := parseLine(line)
	if err != nil {
		return err, nil
	}
	//conn.write(500, err.Error())
	//conn.c.Close()
	_ = args

	switch verb {

	case "HELO":
		{
			////conn.handleHELO(args)
		}

	case "EHLO":
		{
			////conn.handleEHLO(args)
		}

	case "MAIL":
		{
			//conn.srv.handleMail(//conn, args)
		}

	case "RCPT":
		{
			//conn.handleRCPT(args)
		}

	case "DATA":
		{
			//conn.handleDATA(args)
		}

	case "RSET":
		{
			//conn.handleRSET(args)
		}

	case "VRFY", "EXPN", "SEND", "SOML", "SAML":
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

		}

	case "NOOP":
		{
			//conn.handleNOOP(args)
		}

	case "QUIT":
		{
			//conn.handleQUIT(args)
		}

	default:
		{
			/*
				f := conn.srv.extension(verb)
				if f == nil {
					log.Printf("    > Command unrecognized: '%s'", verb)
					//conn.write(500, "Command unrecognized")
					break
				}
			*/
			//f(//conn, args)
		}

		/*
			RFC 5321
			The maximum total length of a reply line including the reply code and
			the <CRLF> is 512 octets.  More information may be conveyed through
			multiple-line replies.
		*/

	}

	return nil, nil
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
