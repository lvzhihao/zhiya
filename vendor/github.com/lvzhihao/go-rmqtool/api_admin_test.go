package rmqtool

import "testing"

func TestAPIClientClusterName(t *testing.T) {
	ret, err := GenerateTestClient().ClusterName()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
}

func TestAPIClientChangeClusterName(t *testing.T) {
	params := map[string]interface{}{
		"name": "rabbit@test-rabbit-changed",
	}
	err := GenerateTestClient().ChangeClusterName(params)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("ChangeClusterName Success")
	}
}

func TestAPIListVhosts(t *testing.T) {
	list, err := GenerateTestClient().ListVhosts()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Vhosts List: ", list)
	}
}

func TestAPIVhost(t *testing.T) {
	err := GenerateTestClient().CreateVhost("test", false)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Vhost Created Success")
		vhost, err := GenerateTestClient().Vhost("test")
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log(vhost)
			err = GenerateTestClient().DeleteVhost("test")
			if err != nil {
				t.Fatal(err)
			} else {
				t.Log("Delete Vhost Successed: ", vhost)
			}
		}
	}
}

func TestAPIListUsers(t *testing.T) {
	list, err := GenerateTestClient().ListUsers()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Users List: ", list)
	}
}

func TestAPIGenerateUserPasswordHash(t *testing.T) {
	passwd := "test12"
	t.Log("org: ", passwd)
	passwdHash := GenerateTestClient().GenerateUserPasswordHash(passwd)
	t.Log("hash: ", passwdHash)
	ok := GenerateTestClient().CheckUserPasswordHash(passwd, passwdHash)
	if ok {
		t.Log("check hash success")
	} else {
		t.Fatal("check hash failure")
	}
}

func TestAPIUser(t *testing.T) {
	err := GenerateTestClient().CreateUser("test", map[string]interface{}{
		"password_hash": GenerateTestClient().GenerateUserPasswordHash("test12"),
		"tags":          "administrator",
	})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("User Created Success")
		vhost, err := GenerateTestClient().User("test")
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log(vhost)
			err = GenerateTestClient().DeleteUser("test")
			if err != nil {
				t.Fatal(err)
			} else {
				t.Log("Delete User Successed: ", vhost)
			}
		}
	}
}

func TestAPIWhoAmI(t *testing.T) {
	user, err := GenerateTestClient().WhoAmI()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Current User: ", user)
	}
}

func TestAPIListPermissions(t *testing.T) {
	list, err := GenerateTestClient().ListPermissions()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Permissions List: ", list)
	}
}

func TestPermissions(t *testing.T) {
	err := GenerateTestClient().CreateUser("test", map[string]interface{}{
		"password_hash": GenerateTestClient().GenerateUserPasswordHash("test12"),
		"tags":          "administrator",
	})
	if err != nil {
		t.Fatal(err)
	} else {
		err := GenerateTestClient().CreatePermission("/", "test", map[string]interface{}{
			"configure": ".*",
			"write":     ".*",
			"read":      ".*",
		})
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log("Create Permissions Success")
		}
		p, err := GenerateTestClient().Permission("/", "test")
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log("Permission: ", p)
		}
		vp, err := GenerateTestClient().VhostPermissions("/")
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log("Vhost Permission List: ", vp)
		}
		up, err := GenerateTestClient().UserPermissions("test")
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log("User Permission List: ", up)
		}
		err = GenerateTestClient().DeletePermission("/", "test")
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log("Delete Permission Success: ", p)
		}
		GenerateTestClient().DeleteUser("test")
	}
}

func TestParameter(t *testing.T) {
	//  only if the federation plugin is enabled
	return
	err := GenerateTestClient().CreateParameter(
		"federation",
		"/",
		"testname",
		map[string]interface{}{
			"vhost":     "/",
			"component": "federation",
			"name":      "testname",
			"value":     "guest",
		},
	)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Create Parameter Success")
	}
	list, err := GenerateTestClient().ListParameters("", "")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Parameters List: ", list)
	}
}

func TestGlobalParameter(t *testing.T) {
	list, err := GenerateTestClient().ListGlobalParameters()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("GlobalParameters List: ", list)
	}
	err = GenerateTestClient().CreateGlobalParameter("test", map[string]interface{}{
		"value": "valueP",
	})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Create GlobalParameter Success")
		p, err := GenerateTestClient().GlobalParameter("test")
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log(p)
			err := GenerateTestClient().DeleteGlobalParameter("test")
			if err != nil {
				t.Fatal(err)
			} else {
				t.Log("Delete GlobalParameter Success: ", p)
			}
		}
	}
}

func TestPolicies(t *testing.T) {
	list, err := GenerateTestClient().ListPolicies("/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Policies List: ", list)
	}
}

func TestAPIAlivenessTest(t *testing.T) {
	ret, err := GenerateTestClient().AlivenessTest("/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(ret)
	}
	err = GenerateTestClient().CreatePolicy("/", "test", map[string]interface{}{
		"pattern": "^amq.",
		"definition": map[string]interface{}{
			"message-ttl": 10,
		},
		"priority": 0,
		"apply-to": "all",
	})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("Create Policy Success")
		p, err := GenerateTestClient().Policy("/", "test")
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log(p)
			err := GenerateTestClient().DeletePolicy("/", "test")
			if err != nil {
				t.Fatal(err)
			} else {
				t.Log("Delete Policy Success: ", p)
			}
		}
	}
}

func TestAPIHealthCheck(t *testing.T) {
	t.Log(GenerateTestClient().HealthCheck())
}
