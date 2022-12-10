package nodebuilder

import (
	"os"
	"path/filepath"

	"github.com/celestiaorg/celestia-node/libs/fslock"
	"github.com/celestiaorg/celestia-node/libs/utils"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
)

// Init initializes the Node FileSystem Store for the given Node Type 'tp' in the directory under
// 'path'.
func Init(cfg Config, path string, tp node.Type) error {
	path, err := storePath(path)
	if err != nil {
		return err
	}
	log.Infof("Initializing %s Node Store over '%s'", tp, path)

	err = initRoot(path)
	if err != nil {
		return err
	}

	flock, err := fslock.Lock(lockPath(path))
	if err != nil {
		if err == fslock.ErrLocked {
			return ErrOpened
		}
		return err
	}
	defer flock.Unlock() //nolint: errcheck

	err = initDir(keysPath(path))
	if err != nil {
		return err
	}

	err = initDir(dataPath(path))
	if err != nil {
		return err
	}

	cfgPath := configPath(path)
	err = SaveConfig(cfgPath, &cfg)
	if err != nil {
		return err
	}
	log.Infow("Saving config", "path", cfgPath)
	log.Info("Node Store initialized")
	return nil
}

func Remove(path string) error {
	strPath, err := storePath(path)
	if err != nil {
		return err
	}

	cfgPath := configPath(strPath)
	return RemoveConfig(cfgPath)
}

func Reinit(cfg Config, path string, newConfigPath string, tp node.Type) error {
	path, err := storePath(path)
	if err != nil {
		return err
	}
	cfgPath := configPath(path)
	exist := utils.Exists(cfgPath)

	// if config exist make changes
	if exist {
		log.Infof("Reinitializing %s Node Store Config over '%s'", tp, path)

		flock, err := fslock.Lock(lockPath(path))
		if err != nil {
			if err == fslock.ErrLocked {
				return ErrOpened
			}
			return err
		}
		defer flock.Unlock() //nolint: errcheck

		newConf, err := LoadConfig(newConfigPath)
		if err != nil {
			return err
		}

		err = SaveConfig(cfgPath, newConf)
		if err != nil {
			return err
		}

		log.Infow("Saving config", "path", cfgPath)
		return nil
	}

	log.Infof("%s Node store not found '%s'", tp, path)
	return Init(cfg, path, tp)
}

// IsInit checks whether FileSystem Store was setup under given 'path'.
// If any required file/subdirectory does not exist, then false is reported.
func IsInit(path string) bool {
	path, err := storePath(path)
	if err != nil {
		log.Errorw("parsing store path", "path", path, "err", err)
		return false
	}

	_, err = LoadConfig(configPath(path)) // load the Config and implicitly check for its existence
	if err != nil {
		log.Errorw("loading config", "path", path, "err", err)
		return false
	}

	if utils.Exists(keysPath(path)) &&
		utils.Exists(dataPath(path)) {
		return true
	}

	return false
}

const perms = 0755

// initRoot initializes(creates) directory if not created and check if it is writable
func initRoot(path string) error {
	err := initDir(path)
	if err != nil {
		return err
	}

	// check for writing permissions
	f, err := os.Create(filepath.Join(path, ".check"))
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return os.Remove(f.Name())
}

// initDir creates a dir if not exist
func initDir(path string) error {
	if utils.Exists(path) {
		return nil
	}
	return os.Mkdir(path, perms)
}
