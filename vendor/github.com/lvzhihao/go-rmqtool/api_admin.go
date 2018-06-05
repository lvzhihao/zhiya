package rmqtool

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
)

func (c *APIClient) ClusterName() (map[string]interface{}, error) {
	return c.readMap("cluster-name")
}

func APIClusterName(api, user, passwd string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ClusterName()
}

func (c *APIClient) ChangeClusterName(params map[string]interface{}) error {
	return c.create("/cluster-name", params)
}

func APIChnageClusterName(api, user, passwd string, params map[string]interface{}) error {
	return NewAPIClient(api, user, passwd).ChangeClusterName(params)
}

func (c *APIClient) ListVhosts() ([]map[string]interface{}, error) {
	return c.readSlice("vhosts")
}

func APIListVhosts(api, user, passwd string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListVhosts()
}

func (c *APIClient) CreateVhost(name string, tracing bool) error {
	return c.create([]string{"vhosts", name}, map[string]bool{
		"tracing": tracing,
	})
}

func APICreateVhost(api, user, passwd, name string, tracing bool) error {
	return NewAPIClient(api, user, passwd).CreateVhost(name, tracing)
}

func (c *APIClient) Vhost(name string) (map[string]interface{}, error) {
	return c.readMap([]string{"vhosts", name})
}

func APIVhost(api, user, passwd, name string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Vhost(name)
}

func (c *APIClient) DeleteVhost(name string) error {
	return c.delete([]string{"vhosts", name})
}

func APIDeleteVhost(api, user, passwd, name string) error {
	return NewAPIClient(api, user, passwd).DeleteVhost(name)
}

func (c *APIClient) VhostPermissions(name string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"vhosts", name, "permissions"})
}

func APIVhostPermissions(api, user, passwd, name string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).VhostPermissions(name)
}

func (c *APIClient) ListUsers() ([]map[string]interface{}, error) {
	return c.readSlice("users")
}

func APIListUsers(api, user, passwd string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListUsers()
}

func (c *APIClient) GenerateUserPasswordHash(passwd string) string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, rand.Uint32())
	salt := buf.Bytes()
	//salt := []byte{144, 141, 198, 10}
	//log.Printf("% x\n", salt)
	sum := sha256.Sum256(append(salt, []byte(passwd)...))
	//log.Printf("% x\n", sum)
	//log.Printf("% x\n", append(salt, sum[:]...))
	return base64.StdEncoding.EncodeToString(append(salt, sum[:]...))
}

func (c *APIClient) CheckUserPasswordHash(passwd, hash string) bool {
	decoded, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false
	}
	//log.Printf("% x\n", decoded[4:])
	salt := append([]byte{}, decoded[0:4]...)
	sum := sha256.Sum256(append(salt, []byte(passwd)...))
	//log.Printf("% x\n", sum)
	//log.Printf("% x\n", decoded[4:])
	if bytes.Compare(sum[:], decoded[4:]) == 0 {
		return true
	} else {
		return false
	}
}

func (c *APIClient) CreateUser(name string, data map[string]interface{}) error {
	return c.create([]string{"users", name}, data)
}

func APICreateUser(api, user, passwd, name string, data map[string]interface{}) error {
	return NewAPIClient(api, user, passwd).CreateUser(name, data)
}

func (c *APIClient) User(name string) (map[string]interface{}, error) {
	return c.readMap([]string{"users", name})
}

func APIUser(api, user, passwd, name string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).User(name)
}

func (c *APIClient) DeleteUser(name string) error {
	return c.delete([]string{"users", name})
}

func APIDeleteUser(api, user, passwd, name string) error {
	return NewAPIClient(api, user, passwd).DeleteUser(name)
}

func (c *APIClient) UserPermissions(name string) ([]map[string]interface{}, error) {
	return c.readSlice([]string{"users", name, "permissions"})
}

func APIUserPermissions(api, user, passwd, name string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).UserPermissions(name)
}

func (c *APIClient) WhoAmI() (map[string]interface{}, error) {
	return c.readMap("whoami")
}

func APIWhoAmI(api, user, passwd string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).WhoAmI()
}

func (c *APIClient) ListPermissions() ([]map[string]interface{}, error) {
	return c.readSlice("permissions")
}

func APIListPermissions(api, user, passwd string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListPermissions()
}

func (c *APIClient) CreatePermission(vhost, user string, data map[string]interface{}) error {
	return c.create([]string{"permissions", vhost, user}, data)
}

func APICreatePermission(api, user, passwd, vhost, username string, data map[string]interface{}) error {
	return NewAPIClient(api, user, passwd).CreatePermission(vhost, username, data)
}

func (c *APIClient) Permission(vhost, user string) (map[string]interface{}, error) {
	return c.readMap([]string{"permissions", vhost, user})
}

func APIPermission(api, user, passwd, vhost, username string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Permission(vhost, username)
}

func (c *APIClient) DeletePermission(vhost, user string) error {
	return c.delete([]string{"permissions", vhost, user})
}

func APIDeletePermission(api, user, passwd, vhost, username string) error {
	return NewAPIClient(api, user, passwd).DeletePermission(vhost, username)
}

func (c *APIClient) ListParameters(component, vhost string) ([]map[string]interface{}, error) {
	if component != "" && vhost != "" {
		return c.readSlice([]string{"parameters", component, vhost})
	} else if component != "" {
		return c.readSlice([]string{"parameters", component})
	} else {
		return c.readSlice([]string{"parameters"})
	}
}

func APIListParameters(api, user, passwd, component, vhost string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListParameters(component, vhost)
}

func (c *APIClient) Parameter(component, vhost, name string) (map[string]interface{}, error) {
	return c.readMap([]string{"parameters", component, vhost, name})
}

func APIParameter(api, user, passwd, component, vhost, pname string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Parameter(component, vhost, pname)
}

func (c *APIClient) CreateParameter(component, vhost, name string, data map[string]interface{}) error {
	return c.create([]string{"parameters", component, vhost, name}, data)
}

func APICreateParameter(api, user, passwd, component, vhost, pname string, data map[string]interface{}) error {
	return NewAPIClient(api, user, passwd).CreateParameter(component, vhost, pname, data)
}

func (c *APIClient) DeleteParameter(component, vhost, name string) error {
	return c.delete([]string{"parameters", component, vhost, name})
}

func APIDeleteParameter(api, user, passwd, component, vhost, name string) error {
	return NewAPIClient(api, user, passwd).DeleteParameter(component, vhost, name)
}

func (c *APIClient) ListGlobalParameters() ([]map[string]interface{}, error) {
	return c.readSlice("global-parameters")
}

func APIListGlobalParameters(api, user, passwd string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListGlobalParameters()
}

func (c *APIClient) GlobalParameter(name string) (map[string]interface{}, error) {
	return c.readMap([]string{"global-parameters", name})
}

func APIGlobalParameter(api, user, passwd, name string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).GlobalParameter(name)
}

func (c *APIClient) CreateGlobalParameter(name string, data map[string]interface{}) error {
	return c.create([]string{"global-parameters", name}, data)
}

func APICreateGlobalParameter(api, user, passwd, name string, data map[string]interface{}) error {
	return NewAPIClient(api, user, passwd).CreateGlobalParameter(name, data)
}

func (c *APIClient) DeleteGlobalParameter(name string) error {
	return c.delete([]string{"global-parameters", name})
}

func APIDeleteGlobalParameter(api, user, passwd, name string) error {
	return NewAPIClient(api, user, passwd).DeleteGlobalParameter(name)
}

func (c *APIClient) ListPolicies(vhost string) ([]map[string]interface{}, error) {
	if vhost == "" {
		return c.readSlice("policies")
	} else {
		return c.readSlice([]string{"policies", vhost})
	}
}

func APIListPolicies(api, user, passwd, vhost string) ([]map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).ListPolicies(vhost)
}

func (c *APIClient) Policy(vhost, name string) (map[string]interface{}, error) {
	return c.readMap([]string{"policies", vhost, name})
}

func APIPolicy(api, user, passwd, vhost, name string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).Policy(vhost, name)
}

func (c *APIClient) CreatePolicy(vhost, name string, data map[string]interface{}) error {
	return c.create([]string{"policies", vhost, name}, data)
}

func APICreatePolicy(api, user, passwd, vhost, name string, data map[string]interface{}) error {
	return NewAPIClient(api, user, passwd).CreatePolicy(vhost, name, data)
}

func (c *APIClient) DeletePolicy(vhost, name string) error {
	return c.delete([]string{"policies", vhost, name})
}

func APIDeletePolicy(api, user, passwd, vhost, name string) error {
	return NewAPIClient(api, user, passwd).DeletePolicy(vhost, name)
}

func (c *APIClient) AlivenessTest(vhost string) (map[string]interface{}, error) {
	return c.readMap([]string{"aliveness-test", vhost})
}

func APIAlivenessTest(api, user, passwd, vhost string) (map[string]interface{}, error) {
	return NewAPIClient(api, user, passwd).AlivenessTest(vhost)
}

func (c *APIClient) HealthCheck() error {
	return c.NodeHealthCheck("")
}

func APIHealthCheck(api, user, passwd string) error {
	return NewAPIClient(api, user, passwd).HealthCheck()
}

func (c *APIClient) NodeHealthCheck(node string) error {
	var data map[string]interface{}
	var err error
	if node == "" {
		data, err = c.readMap([]string{"healthchecks", "node"}) //self
	} else {
		data, err = c.readMap([]string{"healthchecks", "node", node}) //node
	}
	if err != nil {
		return err
	}
	if status, ok := data["status"]; !ok {
		return fmt.Errorf("undefined error")
	} else {
		if status.(string) == "ok" {
			return nil
		} else {
			return fmt.Errorf("%s", data["reason"])
		}
	}
}

func APINodeHealthCheck(api, user, passwd, node string) error {
	return NewAPIClient(api, user, passwd).NodeHealthCheck(node)
}
