package jschemagomd

import (
	"fmt"
	"log"
	"text/template"

	"golang.org/x/net/html"
)

//InitTemplate ...
func InitTemplate() {
	var err error
	mdTemplate, err = template.ParseGlob("*.gotmpl")
	//remember to remove
	if err != nil {
		log.Printf("error while parsing the template %s %v", "*.gotmpl", err)
		mdTemplate, err = template.ParseFiles("mdown.gotmpl")
	}
}

//ExtractObjectFromJSchema assuming a map has the type as key and the value is object
//if that condition met then properties should exist
func ExtractObjectFromJSchema(itType map[string]interface{}) map[string]*JSchemaField {
	if tyval, ok := itType["type"]; ok {
		if tyval == "object" {

			if prop, exists := itType["properties"]; exists {
				if allProp, isProp := prop.(map[string]interface{}); isProp {
					m := make(map[string]*JSchemaField)
					for k := range allProp {

						if p, pTypeMatch := allProp[k].(map[string]interface{}); pTypeMatch {
							//log.Printf("key %v,type %v", k, p["type"])
							if t, hasType := p["type"]; hasType {
								if tVal, ok := t.(string); ok {
									if tVal == "object" || tVal == "array" {
										j := &JSchemaField{JFType: tVal}
										j.AssignNameDescriptionIfAvailable(p)
										if tVal == "array" {
											j.AddSubField(ExtractArrayFromJSchema(p), GetFieldFromItems("description", p))
										} else {
											j.AddSubField(ExtractObjectFromJSchema(p), "")
										}
										m[k] = j
									} else {
										m[k] = BuildASingleField(p, tVal)
									}

								}
							} else {
								//no type is defined
								//p may contain other attribute
								m[k] = BuildASingleField(p, "")
							}

						}

					}
					return UpdateFieldForRequiredField(itType, m)
				}
			}
		} else {
			//in case it is a primitive type
			log.Printf("It is not necessary to handle %v ", itType)
		}
	} else {
		log.Printf("this field %v is missing the type. Try to check if the properties exist", itType["$id"])
		if prop, propExist := itType["properties"]; propExist {
			log.Println("This is an object check")
			if propVal, propTypeOk := prop.(map[string]interface{}); propTypeOk {
				j := &JSchemaField{JFType: "object"}
				j.AddSubField(ExtractObjectFromJSchema(propVal), "")
			}

		} else {
			log.Println("It is not necessary to check")
		}
	}
	return nil
}

//BuildASingleField ...
func BuildASingleField(p map[string]interface{}, t string) *JSchemaField {
	jf := &JSchemaField{}
	jf.AssignNameDescriptionIfAvailable(p)

	if tyVal, ok := p["type"]; ok {
		//t := aType.(string)
		//the type can be complicated, for instance, instead of string,integer,array,object it is
		//['string','null']
		if actType, ok := tyVal.(string); ok {
			jf.JFType = actType
		} else {
			var sType string
			if actType, ok := tyVal.([]interface{}); ok {
				for i, x := range actType {
					if i == 0 {
						sType = x.(string)
					} else {
						sType = fmt.Sprintf("[%s,%s]", sType, x.(string))
					}

				}
				jf.JFType = sType
			}
		}

		switch t {
		case Str.String():
			jf.AdditionalProperties = CopyMapIfExist(p, "pattern", "examples")
			break
		case Int.String():
			jf.AdditionalProperties = CopyMapIfExist(p, "minimum", "maximum", "default")
			break
		case Obj.String():
			//retrieve object
			//p is map
			//get p["properties"]
			sf := ExtractObjectFromJSchema(p)
			jf.SubField = sf

			break
		case Arr.String():
			sf := ExtractArrayFromJSchema(p)
			jf.SubField = sf
			break
		case BOOL.String():
			jf.AdditionalProperties = CopyMapIfExist(p, "default")
			break

		}
	} else {
		//type is not present, but anyOf
		//it is like dynamic type for some possibilities
		jf.DynamicType = make([]interface{}, 0)
		if arAny, ok := p["anyOf"]; ok {
			if arIn, ok := arAny.([]interface{}); ok {
				for _, ararIn := range arIn {
					jf.DynamicType = append(jf.DynamicType, ararIn)
				}

			}
		} else {
			//just blindly create object with minimum attribute
			jf.AdditionalProperties = make(map[string]interface{})
			for k, v := range p {
				jf.AdditionalProperties[k] = v
			}

		}

	}

	return jf
}

//CopyMapIfExist ...
func CopyMapIfExist(m map[string]interface{}, kTarget ...string) map[string]interface{} {
	newmap := make(map[string]interface{})
	for _, t := range kTarget {
		if k, ok := m[t]; ok {
			if t == "pattern" {
				//convert k into html escape
				newmap[t] = html.EscapeString(k.(string))
			} else {
				if t == "examples" {
					if arr, isArr := k.([]interface{}); isArr {
						var ex string

						for idx, v := range arr {
							if idx == 0 {
								ex = fmt.Sprintf("%s", v)
							} else {
								ex = fmt.Sprintf("%s,%s", ex, v)
							}

						}
						newmap[t] = ex
					} else {
						log.Printf("key=%s,value=%v", t, k)
						newmap[t] = k
					}
				} else {
					newmap[t] = k
				}

			}

		}
	}
	if len(newmap) > 0 {
		return newmap
	}
	return nil
}

//RetrieveMapValue ...
func RetrieveMapValue(m map[string]interface{}, k string) string {
	if v, ok := m[k]; ok {
		if v != nil {
			return v.(string)
		}
		log.Printf("Nill value key=%s %v", k, m)

	}
	return ""
}

//ExtractArrayFromJSchema assuming a map has the type as key and the value is array
//if that condition met then items should exist
func ExtractArrayFromJSchema(ty map[string]interface{}) map[string]*JSchemaField {
	if ar, arOk := ty["type"]; arOk && ar == "array" {
		if it, itOk := ty["items"]; itOk {
			if itType, itTypeOk := it.(map[string]interface{}); itTypeOk {
				//need to get all the properties
				//Normally, if the type os object then the properties field exists and have the
				//subfield description
				return UpdateFieldForRequiredField(itType, ExtractObjectFromJSchema(itType))
			}

		}
	}
	return nil
}

//UpdateFieldForRequiredField ...
func UpdateFieldForRequiredField(itType map[string]interface{}, allProp map[string]*JSchemaField) map[string]*JSchemaField {
	//get description
	//allProp
	if reqField, reqExist := itType["required"]; reqExist && len(allProp) > 0 {
		//if reqField
		return UpdateRequiredField(reqField, allProp)
	}
	return allProp
}

//UpdateRequiredField ...
func UpdateRequiredField(regField interface{}, allProp map[string]*JSchemaField) map[string]*JSchemaField {
	if reqs, tOk := regField.([]interface{}); tOk {
		//log.Printf("required fields %v", reqs)
		if len(reqs) > 0 {
			for k := range allProp {
				//jf = allProp[k]
				for _, v := range reqs {
					if k == v {
						allProp[k].Required = true
						break
					}
				}
			}
		}

	}
	return allProp
}
