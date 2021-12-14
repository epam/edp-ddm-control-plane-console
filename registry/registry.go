package registry

const (
	ScenarioKeyRequired    = "key-required"
	ScenarioKeyNotRequired = "key-not-required"
	KeyDeviceTypeFile      = "file"
	KeyDeviceTypeHardware  = "hardware"
)

type registry struct {
	Name                string   `form:"name" binding:"required,min=3,max=12,registry-name"`
	Description         string   `form:"description" valid:"max=250"`
	Admins              string   `form:"admins" binding:"registry-admins"`
	SignKeyIssuer       string   `form:"sign-key-issuer" binding:"required_if=KeyDeviceType file Scenario key-required"`
	SignKeyPwd          string   `form:"sign-key-pwd" binding:"required_if=KeyDeviceType file Scenario key-required"`
	RegistryGitTemplate string   `form:"registry-git-template" binding:"required"`
	RegistryGitBranch   string   `form:"registry-git-branch" binding:"required"`
	KeyDeviceType       string   `form:"key-device-type" binding:"oneof=file hardware"`
	RemoteType          string   `form:"remote-type" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyPassword   string   `form:"remote-key-pwd" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAName        string   `form:"remote-ca-name" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAHost        string   `form:"remote-ca-host" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAPort        string   `form:"remote-ca-port" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteSerialNumber  string   `form:"remote-serial-number" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyPort       string   `form:"remote-key-port" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyHost       string   `form:"remote-key-host" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyMask       string   `form:"remote-key-mask" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	Scenario            string   `binding:"oneof=key-required key-not-required"`
	INIConfig           string   `form:"remote-ini-config" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	AllowedKeysSerial   []string `form:"allowed-keys-serial[]" binding:"required_if=Scenario key-required"`
	AllowedKeysIssuer   []string `form:"allowed-keys-issuer[]" binding:"required_if=Scenario key-required"`
}

func (r *registry) KeysRequired() bool {
	return r.Scenario == ScenarioKeyRequired
}

type allowedKeysConfig struct {
	AllowedKeys []allowedKey `yaml:"allowed-keys"`
}

type allowedKey struct {
	Issuer string `yaml:"issuer"`
	Serial string `yaml:"serial"`
}
