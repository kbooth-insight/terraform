package azure_vault

import (
	"bytes"
	"context"
	keyvaultClient "github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"

	//"encoding/base64"
	//"encoding/json"
	"fmt"
	//"io"
	//"log"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2018-02-14/keyvault"
	//"github.com/hashicorp/go-multierror"
	//"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform/state"
	"github.com/hashicorp/terraform/state/remote"
	//"github.com/hashicorp/terraform/states"
)

const (
	leaseHeader = "x-ms-lease-id"
	// Must be lower case
	lockInfoMetaKey = "terraformlockid"
)

type RemoteClient struct {
	vaultClient    keyvault.VaultsClient
	keyvaultName   string
	keyvaultPrefix string

	leaseID       string
	secretsClient keyvaultClient.BaseClient
	keyName       string
}

func (c *RemoteClient) Get() (*remote.Payload, error) {

	ctx := context.TODO()
	secret, e := c.secretsClient.GetSecret(ctx, c.keyvaultName, c.keyvaultPrefix, "")
	if e != nil {
		return nil, fmt.Errorf("Unable to fetch secret: %s", c.keyvaultPrefix)
	}

	fmt.Printf("secret value: %s", secret.Value)

	buf := bytes.NewBuffer(nil)

	payload := &remote.Payload{
		Data: buf.Bytes(),
	}

	// If there was no data, then return nil
	if len(payload.Data) == 0 {
		return nil, nil
	}

	return payload, nil
}

func (c *RemoteClient) Put(data []byte) error {

	return nil
}

func (c *RemoteClient) Delete() error {

	return nil
}

func (c *RemoteClient) Lock(info *state.LockInfo) (string, error) {
	fmt.Printf("Locking state...")
	info.Path = "stateName"

	return info.ID, nil
}

func (c *RemoteClient) getLockInfo() (*state.LockInfo, error) {
	//containerReference := c.blobClient.GetContainerReference(c.containerName)
	//blobReference := containerReference.GetBlobReference(c.keyName)
	//err := blobReference.GetMetadata(&storage.GetBlobMetadataOptions{})
	//if err != nil {
	//	return nil, err
	//}

	//raw := blobReference.Metadata[lockInfoMetaKey]
	//if raw == "" {
	//	return nil, fmt.Errorf("blob metadata %q was empty", lockInfoMetaKey)
	//}

	//data, err := base64.StdEncoding.DecodeString(raw)
	//if err != nil {
	//	return nil, err
	//}

	//lockInfo := &state.LockInfo{}
	//err = json.Unmarshal(data, lockInfo)
	//if err != nil {
	//	return nil, err
	//}

	return nil, nil
}

// writes info to blob meta data, deletes metadata entry if info is nil
func (c *RemoteClient) writeLockInfo(info *state.LockInfo) error {
	//containerReference := c.blobClient.GetContainerReference(c.containerName)
	//blobReference := containerReference.GetBlobReference(c.keyName)
	//err := blobReference.GetMetadata(&storage.GetBlobMetadataOptions{
	//	LeaseID: c.leaseID,
	//})
	//if err != nil {
	//	return err
	//}
	//
	//if info == nil {
	//	delete(blobReference.Metadata, lockInfoMetaKey)
	//} else {
	//	value := base64.StdEncoding.EncodeToString(info.Marshal())
	//	blobReference.Metadata[lockInfoMetaKey] = value
	//}
	//
	//opts := &storage.SetBlobMetadataOptions{
	//	LeaseID: c.leaseID,
	//}
	return nil
}

func (c *RemoteClient) Unlock(id string) error {
	lockErr := &state.LockError{}

	lockInfo, err := c.getLockInfo()
	if err != nil {
		lockErr.Err = fmt.Errorf("failed to retrieve lock info: %s", err)
		return lockErr
	}
	lockErr.Info = lockInfo

	if lockInfo.ID != id {
		lockErr.Err = fmt.Errorf("lock id %q does not match existing lock", id)
		return lockErr
	}

	c.leaseID = lockInfo.ID
	if err := c.writeLockInfo(nil); err != nil {
		lockErr.Err = fmt.Errorf("failed to delete lock info from metadata: %s", err)
		return lockErr
	}

	//containerReference := c.blobClient.GetContainerReference(c.containerName)
	//blobReference := containerReference.GetBlobReference(c.keyName)
	//err = blobReference.ReleaseLease(id, &storage.LeaseOptions{})
	//if err != nil {
	//	lockErr.Err = err
	//	return lockErr
	//}

	c.leaseID = ""

	return nil
}
