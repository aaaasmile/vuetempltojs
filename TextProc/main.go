package main

import (
	"flag"
	"fmt"
	"github/aaaasmile/TextProc/lexer"
	"github/aaaasmile/TextProc/lexerjs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Lo scopo di questo programma è quello di prendere il contenuto del file vue di input e:
// 1) Trovare il contenuto del tab <template>
// 2) Sostituirlo nel file del component in js (stesso nome impleicito ma con estensione .js) vale a dire
//    sostituire il valore tra gli apici dove si trova template: ``
// La ragioni sono diverse. Un file .vue ha il controllo dei colori e dei tag nell'editor. Però mescola script js e css (tema "Separation of Concerns").
// Personalmente piace avere un file con molti tag in un file separato. Il javascript meglio per suo conto e la sezione css proprio non lo voglio vdere.
// Un file vue comunque è sempre meglio di certi framework che per un componente generano 4 files (js, template xml, css e test).
// Ma la motivazione principale è che nello sviluppo del front-end vue servito da un back-end in golang di solito avviene in progetti separati.
// Per sviluppare il frton-end con i file di tipo .vue occorre un build processor, tipicamente webpack e node. Poi quando lo sviluppo è finito
// lo si mette nel backend. La coppia webpack e node è un polpettone impressionante. Ogni volta che parte lancia warning, scarica moduli
// con interdipendenze che dopo qualche mese diventano ingestibili. Per creare un progetto hello-world per l'estensione di visual code,
// node ha prima installato 40Mb di sorgenti in 12800 files. Questo per avere un file json e due js con 40 linee di codice predefinito ababstanza triviali.
// Lo sviluppo del frontend js in un progetto backend in golang, invece, non richiede nessun preprocessor, specialmente se il progetto è molto piccolo.
// Però editare template vue in js nella variabile template: `` è abbastanza penoso. Ecco perchè l'idea di editare il template nel file .vue
// avendo in automatico il risultato nel file js senza avere un preprocessor della mole di node e webpack.

// Command line example: .\TextProc.exe -vue .\example\home.vue
const Buildnr = "00.01.03-20200330"
const LogfileName = "TextProc.log"

func main() {
	log.SetOutput(os.Stdout)
	var vueFile = flag.String("vue", "", "Vue file that contains the component template")
	var logpath = flag.String("logpath", "", fmt.Sprintf("An absolute file path of a directory in which the tool create the log file %s", LogfileName))
	var ver = flag.Bool("ver", false, "Tool version")

	flag.Parse()

	if *ver {
		fmt.Println("TextProc version ", Buildnr)
		os.Exit(0)
	}

	if *logpath != "" {
		if _, err := os.Stat(*logpath); os.IsNotExist(err) {
			os.Mkdir(*logpath, 0777)
		}
		fname := filepath.Join(*logpath, LogfileName)
		f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		log.Println("Log file is ", fname)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	if *vueFile == "" {
		log.Fatal("Vue file is empty")
	}
	if filepath.Ext(*vueFile) != ".vue" {
		log.Fatalf("The file %s is not a vue file", *vueFile)
	}

	vt := lexer.VueTempl{} // uso un lexer e non xml UnmarshalXML in quanto il contenuto del file vue non si lascia scansionare con un parser xml puro
	tmpl, err := vt.GetTemplateFromFile(*vueFile)
	if err != nil {
		log.Fatalln("Error on processing vue file: ", err)
	}
	log.Println("Template content is", tmpl)

	dir, file := filepath.Split(*vueFile)
	fileJs := strings.TrimSuffix(file, filepath.Ext(file))
	fileJs = filepath.Join(dir, fileJs+".js") // file di destinazione con lo stesso nome ma estensione .js
	if _, err := os.Stat(fileJs); err != nil {
		log.Fatalf("Destination file %s not found", fileJs)
	}
	log.Println("Prepare to update the file ", fileJs)
	jt := lexerjs.JsTempl{}
	err = jt.SplitContentComponent(fileJs)
	if err != nil {
		log.Fatalln("Error on processing js file", err)
	}
	//fmt.Printf("*** info\n%s", jt.String())
	jt.SectionContent = tmpl
	if err := jt.WriteFile(fileJs); err != nil {
		log.Fatal("Error ", err)
	}
	log.Printf("File %s successfully updated", fileJs)
}
