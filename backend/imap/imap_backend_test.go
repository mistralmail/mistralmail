package imapbackend

// var test backend.User = &IMAPUser{}

/*

func TestBackend(t *testing.T) {
	// Setup phase
	setup(t)
	defer teardown(t)

	// Create backend
	backend, err := NewIMAPBackend(nil)
	require.NoError(t, err, "couldn't create backend")
	require.NotNil(t, backend)

	// Test login
	user, err := backend.Login(testConnInfo(t), "invalidUsername", "password")
	assert.Error(t, err)
	assert.Nil(t, user)

	user, err = backend.Login(testConnInfo(t), "username", "invalidPassword")
	assert.Error(t, err)
	assert.Nil(t, user)

	user, err = backend.Login(testConnInfo(t), "username", "password")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	// Create a new mailbox for the user
	mailboxName := "TestMailbox"
	err = user.CreateMailbox(mailboxName)
	require.NoError(t, err)

	// Rename the mailbox
	newMailboxName := "RenamedMailbox"
	err = user.RenameMailbox(mailboxName, newMailboxName)
	require.NoError(t, err)

	// Verify that the mailbox was renamed successfully
	renamedMailbox, err := user.GetMailbox(newMailboxName)
	assert.NoError(t, err)
	assert.NotNil(t, renamedMailbox)

	// Verify that the old mailbox name is no longer present
	oldMailbox, err := user.GetMailbox(mailboxName)
	assert.Error(t, err)
	assert.Nil(t, oldMailbox)

	// Delete the new mailbox
	err = user.DeleteMailbox(newMailboxName)
	assert.NoError(t, err)
	assert.NotNil(t, renamedMailbox)

	// Verify that the new mailbox is also no longer present
	_, err = user.GetMailbox(newMailboxName)
	assert.Error(t, err)

}

const dbFile = "./test_backend.db"

// setup function to be called before each test case
func setup(t *testing.T) {

	_, err := os.Stat(dbFile)
	if !os.IsNotExist(err) {
		err = os.Remove(dbFile)
		if err != nil {
			require.NoError(t, err, "couldn't remove db file")
		}
	}

	InitDB(fmt.Sprintf("sqlite:%s", dbFile))
	seedDB()
}

// teardown function to be called after each test case
func teardown(t *testing.T) {
	CloseDB()

	err := os.Remove(dbFile)
	require.NoError(t, err, "couldn't remove db file")
}

func testConnInfo(t *testing.T) *imap.ConnInfo {

	remoteAddr, err := net.ResolveTCPAddr("tcp", "1.1.1.1:143")
	require.NoError(t, err)
	localAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:143")
	require.NoError(t, err)

	return &imap.ConnInfo{
		RemoteAddr: remoteAddr,
		LocalAddr:  localAddr,
	}
}

*/
