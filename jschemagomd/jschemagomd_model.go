package jschemagomd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var (
	mdTemplate   *template.Template
	jsonfileList []string
	//JsonfileDir the directory for the source file
	JsonfileDir string
	//MdfileDir the directory for the result file
	MdfileDir string
	//SchemaNamePattern the pattern for the json schema name
	SchemaNamePattern string
	//PayloadNamePattern the pattern for the json file name
	PayloadNamePattern string
	//ImageFilePattern the pattern for the image file name
	ImageFilePattern string
)

const (
	//Str ...
	Str JSTYPE = "string"
	//Int ...
	Int JSTYPE = "integer"
	//Arr ...
	Arr JSTYPE = "array"
	//Obj ...
	Obj JSTYPE = "object"
	//BOOL ...
	BOOL JSTYPE = "boolean"
	//CurlyOpen ...
	CurlyOpen = "&#123;"
	//CurlyClose ...
	CurlyClose = "&#125;"
	//CONFNAME ...
	CONFNAME = "jschemagomd"
)

func (k JSTYPE) String() string {
	switch k {
	case Str:
		return "string"
	case Int:
		return "integer"
	case Obj:
		return "object"
	case Arr:
		return "array"
	case BOOL:
		return "boolean"
	default:
		return ""
	}

}

//Jschema ...
type Jschema struct {
	SchemaDisplay     string
	SchemaFileName    string
	MdFolder          string
	RequiredFields    []string               `json:"required"`
	SchemaTitle       string                 `json:"title"`
	SchemaDesc        string                 `json:"description"`
	SchemaProperties  map[string]interface{} `json:"properties"`
	SchemaDefinitions map[string]interface{} `json:"definitions"`
	AllFields         map[string]*JSchemaField
	BJschema          []byte
	AllDefinitions    map[string]*JSchemaField
	JSONPayload       string
	Images            map[string]string
}

//JSchemaField is to wrap the properties of a single field
//name, description, value specification, type
type JSchemaField struct {
	JFName               string
	JFDescription        string
	JFType               string //if the type is other than integer or string, then it needs the sub field
	Required             bool
	AdditionalProperties map[string]interface{}
	SubFieldDescription  string
	SubField             map[string]*JSchemaField //array of subfield. This will be an empty slice if the JFType is string or integer
	DynamicType          []interface{}
	RelationConstraints  map[string]interface{}
}

//JSTYPE ...
type JSTYPE string

//JSchemaFieldValue is to wrap value specification
type JSchemaFieldValue struct {
	ValueDesc string
}

//MdizeF ...
type MdizeF struct {
	src  string
	dest string
}

//Copy ...
func (m *MdizeF) Copy() {
	srcStat, err := os.Stat(m.src)
	if err != nil {
		log.Printf("error while starting the stat for %s %v", m.src, err)
	}
	if !srcStat.Mode().IsRegular() {
		ee := fmt.Errorf("%s is not a regular file", m.src)
		log.Printf("error while starting the stat for %s %v", m.src, ee)
	}
	src, errOpen := os.Open(m.src)
	if errOpen != nil {
		log.Printf("error while Opening %s %v", m.src, errOpen)
	}
	defer src.Close()

	dest, err := os.Create(m.dest)
	if err != nil {
		log.Printf("error while creating destination %s %v", m.dest, err)
	}
	defer dest.Close()

	if nbytes, err := io.Copy(dest, src); err != nil {
		log.Printf("error while copying from %s to %s %v", m.src, m.dest, err)
	} else {
		log.Printf("Copied %d bytes from %s to %s ", nbytes, m.src, m.dest)
	}

}

//CreateMD ...
func (j *Jschema) CreateMD() {
	mdName := fmt.Sprintf("%s/%s.md", j.MdFolder, j.SchemaFileName)
	f, eropenFile := os.Create(mdName)
	if eropenFile != nil {
		log.Printf("Error while creating md file %v", eropenFile)

	}
	if err := mdTemplate.ExecuteTemplate(f, "md.gotmpl", j); err != nil {
		if erCl := f.Close(); erCl != nil {
			log.Fatalf("Error while closing the md file %v", erCl)
		}
		log.Printf("error while executing template %s for json schema %s, %v", "md.gotmpl", j.SchemaFileName, err)
	}
	if erCl := f.Close(); erCl != nil {
		log.Printf("Error while closing the md file %v", erCl)
	}
	//t.Execute(os.Stdout, r)
}

//DirWalkerForJSONSchema ...
func DirWalkerForJSONSchema(path string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		//files = append(files, path)
		f := filepath.Base(path)
		if match, err := regexp.MatchString(SchemaNamePattern, f); err == nil {
			if match {
				log.Printf("%s is match %s pattern", f, SchemaNamePattern)
				processJSONSchema(path)
			}
		} else {
			log.Printf("Ignore the file as it is not a json %s", f)
		}

	}
	return nil
}

//processJSONSchema ...
func processJSONSchema(filePath string) {
	fName := filepath.Base(filePath)
	fNameNoext := fName[0 : len(fName)-len(filepath.Ext(fName))]
	fName = fNameNoext[0 : len(fNameNoext)-len(filepath.Ext(fNameNoext))]
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error while reading the file %s %v", filePath, err)
	}
	//mdizer.InitTemplate()
	js := Jschema{SchemaFileName: fName, MdFolder: MdfileDir}
	js.Read(b)
	js.Construct()
	js.Customize()
	js.ConstructDefinition()

	log.Printf("Processing %s ", js.SchemaTitle)
	if len(js.AllFields) == 0 {
		log.Printf("something wrong %v ", js.SchemaTitle)
	} else {
		for k := range js.AllFields {
			log.Printf("key=%s", k)
		}
	}
	js.ReadJSONPayload(filePath)
	js.CollectAllImageNames(filepath.Dir(filePath), filepath.Base(filePath))
	log.Printf("completed %s ", js.SchemaTitle)
	js.CreateMD()
}

//ListFiles ...
func ListFiles(dir string, patternName string) []string {
	files := make([]string, 0)
	filepath.Walk(dir, func(absPath string, fInfo os.FileInfo, err error) error {
		if !fInfo.IsDir() {
			fb := filepath.Base(absPath)
			if matched, errMatching := regexp.MatchString(patternName, fb); errMatching == nil {
				if matched {
					files = append(files, absPath)
				}
			} else {
				log.Fatalf("Error while matching %s %v ", absPath, errMatching)
				return errMatching
			}
		}
		return nil
	})
	return files
}

func getFileWithoutExtension(fileName string) string {

	return fileName[0 : len(fileName)-len(filepath.Ext(fileName))]
}

//ReadJSONPayload ...
func (j *Jschema) ReadJSONPayload(path string) {
	file := filepath.Base(path)
	file = file[0 : len(file)-len(filepath.Ext(file))]
	dir := filepath.Dir(path)
	//build a regex pattern
	fregex := strings.ReplaceAll(file, ".", "\\.")
	plpattern := fmt.Sprintf(PayloadNamePattern, fregex)
	log.Printf("payloadNamePattern %s ", PayloadNamePattern)
	//payloadNamePattern = fmt.Sprintf(payloadNamePattern, file)
	//r := regexp.MustCompile(`(?i)`)
	//payloadFile := fmt.Sprintf("%s_payload.json", file)
	payloadFile := ListFiles(dir, plpattern)

	m := make(map[string]interface{})
	if len(payloadFile) > 0 {
		if b, err := ioutil.ReadFile(payloadFile[0]); err == nil {
			if errUn := json.Unmarshal(b, &m); errUn == nil {
				if byJSON, marshallingError := json.MarshalIndent(m, "", "	"); marshallingError == nil {
					//just stringize the byte array
					j.JSONPayload = string(byJSON)
				}
			} else {
				log.Printf("Error while retrieving payload %s %v", payloadFile, errUn)
			}
		} else {
			log.Printf("Error while read payload %s %v", payloadFile, err)
		}
	} else {
		log.Printf("No payload file is given for json schema %s", path)
	}

}

//CollectAllImageNames ...
func (j *Jschema) CollectAllImageNames(rootDir string, filePrefix string) {

	//build a regex pattern
	fName := filepath.Base(filePrefix)
	fNameNoext := fName[0 : len(fName)-len(filepath.Ext(fName))]
	fNameNoext = strings.ReplaceAll(fNameNoext, ".", "\\.")
	imgPattern := fmt.Sprintf(ImageFilePattern, fNameNoext)
	//imageFilePattern = fmt.Sprintf(imageFilePattern, filePrefix)

	imagelist := ListFiles(rootDir, imgPattern)
	if len(imagelist) > 0 {
		j.Images = make(map[string]string)
		for _, im := range imagelist {
			f := filepath.Base(im)
			j.Images[getFileWithoutExtension(f)] = f

		}
	} else {
		log.Printf("No images are provided for json schema %s ", filePrefix)
	}
	//copy files if necessary

	if MdfileDir != JsonfileDir {
		for _, im := range imagelist {
			destFile := filepath.Join(MdfileDir, filepath.Base(im))
			mf := MdizeF{src: im, dest: destFile}
			mf.Copy()
		}
	}

}

//Read ...
func (j *Jschema) Read(jInBytes []byte) {

	if err := json.Unmarshal(jInBytes, j); err != nil {
		log.Fatalf("Error while unmarshalling %v", err)
	}
	j.BJschema = jInBytes
	//j.SchemaDisplay = string(jInBytes)
}

//Customize ...
func (j *Jschema) Customize() {
	m := make(map[string]interface{})
	var unMarshallingErr error
	var marshallingError error
	var byJSON []byte
	if unMarshallingErr = json.Unmarshal(j.BJschema, &m); unMarshallingErr == nil {
		if _, fExist := m["form"]; fExist {
			delete(m, "form")

		}
		if _, fExist := m["form"]; !fExist {
			log.Println("The form key is deleted")

		}
		if byJSON, marshallingError = json.MarshalIndent(m, "", "	"); marshallingError == nil {
			//just stringize the byte array
			j.SchemaDisplay = string(byJSON)
		}
	} else {
		log.Printf("Error while unmarshalling json into map %v", unMarshallingErr)
	}

}

//Construct ...
//assuming the json schema is like
//{
//  "$schema": "http://json-schema.org/draft-04/schema#",
//  "$id": "MmsxcpValidation",
//  "title": "MMSXCP",
//  "description": "MMSSUCP / MMSICP (MMS User/Issuer Connectivity Parameter). This EF contains values for Multimedia Messaging Connectivity Parameters as determined by the issuer, which can be used by the ME for MMS network connection.",
//  "type": "object",
//  "required": ["req_var"],
//  "properties": {
//    "req_var": {
//        "type": "array|object"
//        "(items|properties)" : {
//
//         }
//    }
//   }
//}
func (j *Jschema) Construct() {
	//var allReqs []string

	if &j.SchemaProperties != nil {
		//iterate over the element
		//few possibilities, object,array
		for k, v := range j.SchemaProperties {
			//make sure the type
			//handle accordingly
			//this is the list of required value
			j.ReadMap(k, v,
				func(j *Jschema, k string, jsf *JSchemaField) {
					j.AddNewProperties(k, jsf)
				})
		}
		j.UpdateCompulsoryField(func(j *Jschema, k string, b bool) {
			j.AllFields[k].Required = b

		})
	}
}

//ConstructDefinition ...
func (j *Jschema) ConstructDefinition() {
	if &j.SchemaDefinitions != nil {
		//m := j.SchemaProperties
		//iterate over the element
		//few possibilities, object,array
		for k, v := range j.SchemaDefinitions {
			//make sure the type
			//handle accordingly
			//this is the list of required value
			j.ReadMap(k, v, func(j *Jschema, k string, jsf *JSchemaField) {
				j.AddNewDefinition(k, jsf)
			})
		}
		j.UpdateCompulsoryField(func(j *Jschema, k string, b bool) {
			if j.AllDefinitions != nil {
				j.AllDefinitions[k].Required = b
			}

		})
	}
}

//GetFieldFromItems extract the description for an array
func GetFieldFromItems(fieldName string, arrFields map[string]interface{}) string {
	if its, itemsExist := arrFields["items"]; itemsExist {
		if itsVal, itsTypeOk := its.(map[string]interface{}); itsTypeOk {
			if f, fExists := itsVal[fieldName]; fExists {
				return f.(string)
			}
		}

	}

	return ""
}

//ReadMap ...
func (j *Jschema) ReadMap(k string, v interface{}, fn func(j *Jschema, k string, jsf *JSchemaField)) {
	if ty, typeMatch := v.(map[string]interface{}); typeMatch {
		s := &JSchemaField{}
		//description and title are not always available
		s.AssignNameDescriptionIfAvailable(ty)
		if t, ok := ty["type"].(string); ok {
			s.JFType = t
			if t == "array" || t == "object" {
				if desc, descExist := ty["description"]; descExist {
					s.JFDescription = desc.(string)
				} else {
					if title, titleExist := ty["title"]; titleExist {
						s.JFDescription = title.(string)
					}
				}
				if t == "array" {

					//s.JFDescription=

					s.AddSubField(ExtractArrayFromJSchema(ty), GetFieldFromItems("description", ty))

				} else {
					s.AddSubField(ExtractObjectFromJSchema(ty), "")
					//check if there is anyOf/anyOne
					if anyKeyword, anyExists := ty["anyOne"]; anyExists {
						if s.RelationConstraints == nil {
							s.RelationConstraints = make(map[string]interface{})
						}
						s.RelationConstraints["anyOne"] = anyKeyword
					}
					if anyKeyword, anyExists := ty["anyOf"]; anyExists {
						if s.RelationConstraints == nil {
							s.RelationConstraints = make(map[string]interface{}, 0)
						}
						s.RelationConstraints["anyOf"] = anyKeyword
					}
				}
				//j.AddNewProperties(k, s)

				fn(j, k, s)
			} else {
				//j.AddNewProperties(k, BuildASingleField(ty, t))
				fn(j, k, BuildASingleField(ty, t))
			}
		} else {
			if t, ok := ty["type"].([]interface{}); ok {
				//support maximum 2 possible types
				for _, actualType := range t {
					if aType, aTypeOk := actualType.(string); aTypeOk && aType != "null" {
						//process accrordingly
						s.HandleTypeAccordingly(k, aType, ty, j, fn)

					}
				}
			}
		}

	}

}

//HandleTypeAccordingly ...
func (f *JSchemaField) HandleTypeAccordingly(k string,
	t string,
	ty map[string]interface{},
	js *Jschema,
	fn func(j *Jschema, k string, jsf *JSchemaField)) {

	f.JFType = t
	if t == "array" || t == "object" {
		if t == "array" {
			f.AddSubField(ExtractArrayFromJSchema(ty), GetFieldFromItems("description", ty))

		} else {

			f.AddSubField(ExtractObjectFromJSchema(ty), "")
		}
		//js.AddNewProperties(k, f)
		fn(js, k, f)
	} else {

		//js.AddNewProperties(k, BuildASingleField(ty, t))
		fn(js, k, BuildASingleField(ty, t))
	}
}

//UpdateCompulsoryField ...
func (j *Jschema) UpdateCompulsoryField(fn func(j *Jschema, k string, b bool)) {
	for k := range j.AllFields {
		for _, v := range j.RequiredFields {
			if v == k {
				//j.AllFields[k].Required = true
				fn(j, k, true)
				break
			}
		}

	}

}

//AssignNameDescriptionIfAvailable ...
func (f *JSchemaField) AssignNameDescriptionIfAvailable(m map[string]interface{}) {
	if title, hasTitle := m["title"]; hasTitle {
		f.JFName = title.(string)
	}
	if desc, hasDesc := m["description"]; hasDesc {
		f.JFDescription = desc.(string)
	}
}

//AddSubField ...
func (f *JSchemaField) AddSubField(m map[string]*JSchemaField, subFieldDesc string) {
	if f.SubField == nil {
		f.SubField = make(map[string]*JSchemaField)
	}
	for k, v := range m {
		f.SubField[k] = v
	}
	f.SubFieldDescription = subFieldDesc
}

//MergeProperties ...
func (j *Jschema) MergeProperties(m map[string]*JSchemaField) {
	if j.AllFields == nil {
		j.AllFields = make(map[string]*JSchemaField)
	}
	for k, v := range m {
		j.AllFields[k] = v
	}

}

//AddNewProperties ...
func (j *Jschema) AddNewProperties(k string, s *JSchemaField) {
	if j.AllFields == nil {
		j.AllFields = make(map[string]*JSchemaField)
	}
	j.AllFields[k] = s
}

//AddNewDefinition ...
func (j *Jschema) AddNewDefinition(k string, s *JSchemaField) {
	if j.AllDefinitions == nil {
		j.AllDefinitions = make(map[string]*JSchemaField)
	}
	j.AllDefinitions[k] = s
}

//Process ...
func (j *Jschema) Process(m map[string]interface{},
	fn func(k string, v map[string]interface{}) map[string]*JSchemaField) map[string]*JSchemaField {

	return nil
}
