package service

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	StorageLocation    = "backup-s3-like-storage-location"
	StorageType        = "backup-s3-like-storage-type"
	StorageCredentials = "backup-s3-like-storage-credentials"
)

type BackupConfig struct {
	StorageLocation    string `form:"storage-location" valid:"Required"`
	StorageType        string `form:"storage-type" valid:"Required"`
	StorageCredentials string `form:"storage-credentials" valid:"Required"`
}

func (s *CodebaseService) GetBackupConfig() (*BackupConfig, error) {
	s.BackupSecretName = "backup-credential"

	secret, err := s.Clients.CoreClient.Secrets(s.Namespace).Get(s.BackupSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get backup config")
	}

	return &BackupConfig{
		StorageCredentials: string(secret.Data[StorageCredentials]),
		StorageType:        string(secret.Data[StorageType]),
		StorageLocation:    string(secret.Data[StorageLocation]),
	}, nil
}

func (s *CodebaseService) SetBackupConfig(conf *BackupConfig) error {
	s.BackupSecretName = "backup-credential"
	if err := s.Clients.CoreClient.Secrets(s.Namespace).Delete(s.BackupSecretName, &metav1.DeleteOptions{}); err != nil && !k8sErrors.IsNotFound(err) {
		return errors.Wrapf(err, "unable to delete secret: %s", s.BackupSecretName)
	}

	if _, err := s.Clients.CoreClient.Secrets(s.Namespace).Create(&v1.Secret{
		Data: map[string][]byte{
			StorageLocation:    []byte(conf.StorageLocation),
			StorageType:        []byte(conf.StorageType),
			StorageCredentials: []byte(conf.StorageCredentials),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.Namespace,
			Name:      s.BackupSecretName,
		},
	}); err != nil {
		return errors.Wrapf(err, "unable to create secret: %s", s.BackupSecretName)
	}

	return nil
}
