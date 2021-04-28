package jschemagomd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func createDocInit(cmd *cobra.Command, args []string) error {
	if &JsonfileDir == nil || &MdfileDir == nil {
		return fmt.Errorf("Either jsFolder or mdFolder is not provided")
	}
	if JsonfileDir == "" || MdfileDir == "" {
		return fmt.Errorf("Either jsFolder or mdFolder is not provided")
	}
	if fabs, err := filepath.Abs(JsonfileDir); err != nil {
		log.Fatalf("While when getting the absolute path for %s %v", JsonfileDir, err)
	} else {
		log.Printf("tool is reading from %s", fabs)
		if _, err := os.Stat(fabs); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("The source folder %s does not exist", fabs)
			}
			return fmt.Errorf("Erorr while analyzing source folder %s %v", fabs, err)
		}
	}
	if fabs, err := filepath.Abs(MdfileDir); err != nil {
		log.Fatalf("While when getting the absolute path for %s %v", MdfileDir, err)
	} else {
		if err := createFolderIfNotExist(fabs); err != nil {
			return err
		}

	}
	return nil
}

func createFolderIfNotExist(fabs string) error {
	if fInfo, err := os.Stat(fabs); err != nil {
		if os.IsNotExist(err) {
			log.Printf("Folder %s, does not exist. lets create it...", fabs)
			if errCr := os.Mkdir(fabs, os.FileMode(1)); errCr != nil {
				log.Printf("failed to create the directory %s, %v", fabs, errCr)
				return fmt.Errorf("failed to create the directory %s, %v", fabs, errCr)
			}
			log.Printf("Successfully create folder %s ", fabs)

		} else {
			return fmt.Errorf("Erorr while analizing folder %s %v", fabs, err)
		}

	} else {
		if !fInfo.IsDir() {
			log.Printf("%s is an existing file. Please use other folder name", fabs)
			return fmt.Errorf("%s is an existing file. Please use other folder name", fabs)
		}
	}
	return nil
}

//JschemaGoMdCmd ...
var JschemaGoMdCmd = &cobra.Command{
	Use:     "jschemagomd",
	Short:   "Create markdown files from all json schemas provided inside a specified folder",
	Long:    `Create markdown files from all json schema provided inside a specified folder.The supporting document, like .png files and json payload example will be processed by the tool as well, once it follow the naming convention`,
	Args:    cobra.OnlyValidArgs,
	PreRunE: createDocInit,
	Run:     CreateDoc,
}

//CreateDoc ...
func CreateDoc(cmd *cobra.Command, args []string) {
	//var files []string
	err := filepath.Walk(JsonfileDir, DirWalkerForJSONSchema)
	if err != nil {
		log.Fatalf("Error while listing all jsonschema file %v", err)
	}

}
