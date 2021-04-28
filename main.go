package main

import (
	"log"

	"github.com/samutayuga/jschemagomd/jschemagomd"
	"github.com/spf13/viper"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	jschemagomd.JschemaGoMdCmd.Example = "jschemagomd -d dir -m mdir"
	jschemagomd.JschemaGoMdCmd.Execute()
}

func init() {
	jschemagomd.JschemaGoMdCmd.Flags().StringVarP(&jschemagomd.MdfileDir, "mdFolder", "m", "", "Specify a directory where the Markdown file output to be stored")
	jschemagomd.JschemaGoMdCmd.Flags().StringVarP(&jschemagomd.JsonfileDir, "jsonFolder", "d", "", "Specify a directory that contains the json schema and the supporting files")
	jschemagomd.JschemaGoMdCmd.MarkFlagRequired("jsonFolder")
	jschemagomd.JschemaGoMdCmd.MarkFlagRequired("mdFolder")
	jschemagomd.InitTemplate()
	viper.SetConfigName(jschemagomd.CONFNAME)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if errRead := viper.ReadInConfig(); errRead != nil {
		log.Fatalf("error while reading config file %s %v", jschemagomd.CONFNAME, errRead)
	}
	jschemagomd.SchemaNamePattern = viper.GetString("jschemagomd.schema-file-pattern")
	jschemagomd.PayloadNamePattern = viper.GetString("jschemagomd.payload-file-pattern")
	jschemagomd.ImageFilePattern = viper.GetString("jschemagomd.image-file-pattern")
}
