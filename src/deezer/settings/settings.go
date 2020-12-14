package settings

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tecnologer/deezer/src/models"
	"github.com/tecnologer/go-secrets"
)

//Settings is struct for settings data
type Settings struct {
	Token *models.AcessToken `json:"token"`
}

var (
	seed            string = ""
	folderName      string = ".go-deezer"
	filePath        string = "settings.json"
	currentPath     string
	currentSettings *Settings
)

func init() {
	var err error
	seed = secrets.GetKeyString("deezer.encrypt_seed")
	currentPath, err = os.Getwd()

	if currentPath == "" || err != nil {
		currentPath = "."
	}

	currentSettings, err = GetSettings()
	if err != nil {
		logrus.WithError(err).Warn("loading settings")
		currentSettings = &Settings{
			Token: &models.AcessToken{},
		}
	}
}

//GetToken returns the stored token
func GetToken() (*models.AcessToken, error) {
	var err error
	if currentSettings == nil {

		currentSettings, err = GetSettings()

		if err != nil {
			return nil, errors.Wrap(err, "get token: getting settings")
		}
	}

	return currentSettings.Token, nil
}

func SetToken(token *models.AcessToken) error {
	currentSettings.Token = token
	return writeFile(currentSettings)
}

//GetSettings returns the settings from file
func GetSettings() (*Settings, error) {
	settings := &Settings{
		Token: &models.AcessToken{},
	}
	data, err := readFile()
	if err != nil {
		return nil, errors.Wrapf(err, "get settings: reading file")
	}

	if len(data) == 0 {
		return settings, nil
	}

	data, err = decrypt(data)
	if err != nil {
		return nil, errors.Wrapf(err, "get settings: decrypting file")
	}

	if len(data) == 0 {
		return settings, nil
	}

	err = json.Unmarshal(data, settings)
	if err != nil {
		return nil, errors.Wrapf(err, "get settings: parsing file to *Settings")
	}

	return settings, nil
}

func readFile() ([]byte, error) {
	path := getPath()

	if !settingsExists(path) {
		return make([]byte, 0), nil
	}

	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, errors.Wrap(err, "error reading settings.json")
	}

	return data, nil
}

func writeFile(data *Settings) error {
	json, err := json.MarshalIndent(data, "", "")
	if err != nil {
		return errors.Wrap(err, "write file: parsing struct to json")
	}

	json, err = encrypt(json)
	if err != nil {
		return errors.Wrap(err, "write file: encrypting json")
	}

	folderPath := getFolderPath()
	if !settingsFolderExists(folderPath) {
		err = os.MkdirAll(folderPath, 0755)
		if err != nil {
			return errors.Wrap(err, "write file: creating path")
		}
	}

	return ioutil.WriteFile(getPath(), json, 0644)
}

func decrypt(data []byte) ([]byte, error) {
	encrKey := getEncrypKey(seed)

	c, err := aes.NewCipher(encrKey)
	if err != nil {
		return nil, errors.Wrap(err, "decrypt: error creating cipher")
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, errors.Wrap(err, "decrypt: error creating Galois Counter Mode")
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("decrypt: invalid NonceSize")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	json, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Wrap(err, "decrypt: error decrypting secret file")
	}

	return json, nil
}

func encrypt(data []byte) ([]byte, error) {
	encrKey := getEncrypKey(seed)
	c, err := aes.NewCipher(encrKey)

	if err != nil {
		return nil, errors.Wrap(err, "encrypt: error creating cipher")
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, errors.Wrap(err, "encrypt: error creating Galois Counter Mode")
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(err, "encrypt: error creating random sequence")
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func getEncrypKey(seed string) []byte {
	key := []byte(seed + currentPath)
	h := sha1.New()
	h.Write(key)
	return h.Sum([]byte("tecnologer20"))
}

func getFolderPath() string {
	homePath := os.Getenv("HOME")
	separator := "/"
	if runtime.GOOS == "windows" {
		separator = "\\"
		homePath += "\\AppData\\Local"
	}

	return fmt.Sprintf("%s%s%s%s", homePath, separator, folderName, separator)
}

func getPath() string {

	return fmt.Sprintf("%s%s", getFolderPath(), filePath)
}

func settingsExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func settingsFolderExists(folderPath string) bool {
	info, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
