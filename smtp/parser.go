package smtp

import (
	"bufio"
	"log"
)

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

	var address *MailAddress
	verb, args, err := parseLine(br)
	log.Printf("Verb: %s. args: %v", verb, args)
	if err != nil {
		return nil, err
	}
	//conn.write(500, err.Error())
	//conn.c.Close()

	switch verb {

	case "HELO":
		{
			if len(args) != 1 {
				command = InvalidCmd{Cmd: "HELO", Info: "HELO requires exactly one valid domain"}
				break
			}
			domain := ""
			for _, arg := range args {
				domain = arg.Key
			}
			command = HeloCmd{Domain: domain}
		}

	case "EHLO":
		{
			if len(args) != 1 {
				command = InvalidCmd{Cmd: "EHLO", Info: "EHLO requires exactly one valid address"}
				break
			}
			domain := ""
			for _, arg := range args {
				domain = arg.Key
			}
			command = EhloCmd{Domain: domain}
		}

	case "MAIL":
		{
			fromArg := args["FROM"]
			address, err = parseFROM(fromArg.Key + fromArg.Operator + fromArg.Value)
			if err != nil {
				command = InvalidCmd{Cmd: verb, Info: err.Error()}
				err = nil
			} else {
				command = MailCmd{From: address}
			}
		}

	case "RCPT":
		{
			toArg := args["TO"]
			address, err = parseTO(toArg.Key + toArg.Operator + toArg.Value)
			if err != nil {
				command = InvalidCmd{Cmd: verb, Info: err.Error()}
				err = nil
			} else {
				command = RcptCmd{To: address}
			}
		}

	case "DATA":
		{
			// TODO: write tests for this
			command = DataCmd{
				R: *NewDataReader(br),
			}
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
			user := ""
			for _, arg := range args {
				user = arg.Key
			}
			command = VrfyCmd{Param: user}
		}

	case "EXPN":
		{
			listName := ""
			for _, arg := range args {
				listName = arg.Key
			}
			command = ExpnCmd{ListName: listName}
		}

	case "NOOP":
		{
			command = NoopCmd{}
		}

	case "QUIT":
		{
			command = QuitCmd{}
		}

	case "STARTTLS":
		{
			command = StartTlsCmd{}
		}

	default:
		{
			// TODO: CLEAN THIS UP
			command = UnknownCmd{Cmd: verb, Line: strings.TrimSuffix(verb, "\n")}
		}

	}

	return
}

type Argument struct {
	Key      string
	Value    string
	Operator string
}

// parseLine returns the verb of the line and a list of all comma separated arguments
func parseLine(br *bufio.Reader) (string, map[string]Argument, error) {
	/*
		RFC 5321
		4.5.3.1.4.  Command Line

		The maximum total length of a command line including the command word
		and the <CRLF> is 512 octets.  SMTP extensions may be used to
		increase this limit.
	*/
	buffer, err := ReadUntill('\n', MAX_CMD_LINE, br)
	if err != nil {
		if err == ErrLtl {
			SkipTillNewline(br)
			return string(buffer), map[string]Argument{}, err
		}

		return string(buffer), map[string]Argument{}, err
	}
	line := string(buffer)
	verb := ""
	argMap := map[string]Argument{}

	// Strip \n and \r
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	i := strings.Index(line, " ")
	if i == -1 {
		verb = strings.ToUpper(line)
		return verb, map[string]Argument{}, nil
	}

	verb = strings.ToUpper(line[:i])
	line = line[i+1:]

	tmpArgs := strings.Split(line, " ")
	for _, arg := range tmpArgs {
		argument := Argument{}
		i = strings.IndexAny(arg, ":=")
		if i == -1 {
			argument.Key = strings.TrimSpace(arg)
		} else {
			argument.Key = strings.ToUpper(strings.TrimSpace(arg[:i]))
			argument.Value = strings.TrimSpace(arg[i+1:])
			argument.Operator = arg[i : i+1]
		}

		if len(argument.Key) == 0 {
			continue
		}

		argMap[argument.Key] = argument
	}

	return verb, argMap, nil
}

func parseFROM(from string) (*MailAddress, error) {
	index := strings.Index(from, ":")
	if index == -1 {
		return nil, errors.New("No FROM given (didn't find ':')")
	}
	if strings.ToLower(from[0:index]) != "from" {
		return nil, errors.New("No FROM given")
	}

	address_str := from[index+1:]

	address, err := ParseAddress(address_str)
	if err != nil {
		return nil, err
	}
	return &address, nil
}

func parseTO(to string) (*MailAddress, error) {
	index := strings.Index(to, ":")
	if index == -1 {
		return nil, errors.New("No TO given (didn't find ':')")
	}
	if strings.ToLower(to[0:index]) != "to" {
		return nil, errors.New("No TO given")
	}

	address_str := to[index+1:]

	address, err := ParseAddress(address_str)
	if err != nil {
		return nil, err
	}
	return &address, nil
}
