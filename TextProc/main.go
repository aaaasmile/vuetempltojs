package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"github/aaaasmile/TextProc/lexer"
	"log"
)

type VueStruct struct {
	XMLName  xml.Name `xml:"template"`
	Template string
}

func (v *VueStruct) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		tok, err := d.Token()
		if tok == nil {
			break
		}
		if err != nil {
			return err
		}
		switch se := tok.(type) {
		case xml.CharData:
			label := string(se)
			v.Template = label
			fmt.Println("*** ", label)
		case xml.EndElement:
		}
	}
	return nil
}

func main() {
	var vueFile = flag.String("vue", "", "Vue file with only the template")
	flag.Parse()

	if *vueFile == "" {
		log.Fatal("Vue file is empty")
	}
	// rawFile, err := ioutil.ReadFile(*vueFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// env := VueStruct{}

	// err = xml.Unmarshal(rawFile, &env)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("*** Template is", env)
	vt := lexer.VueTempl{TokenTag: "<template>"}
	tmpl, err := vt.GetTemplateFromFile(*vueFile)
	if err != nil {
		log.Fatalln("Error on processing vue file", err)
	}
	fmt.Println("*** Template is", tmpl)

}
