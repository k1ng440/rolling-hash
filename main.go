package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/k1ng440/rolling-hash/pkg/delta"
	"github.com/k1ng440/rolling-hash/pkg/files"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()		
		return
	}

	switch mode := strings.ToLower(os.Args[1]); mode {
	case "signature", "sig":
		if len(os.Args) != 4 {
			printHelp()
			return 
		}

		arg := os.Args[2:]

		oldFile, err := files.ReadFile(arg[0])
		if err != nil {
			panic(err)
		}

		sigs, err := delta.GenerateSignatures(oldFile, delta.DefaultBlockSize)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = files.WriteSignaturesToFile(arg[1], sigs)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "delta":
		if len(os.Args) != 5 {
			printHelp()
			return 
		}
		arg := os.Args[2:]

		sigs, err := files.ReadSignaturesFromFile(arg[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		newFile, err := files.ReadFile(arg[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		deltas, err := delta.GenerateDelta(newFile, delta.DefaultBlockSize, sigs)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		files.WriteDelta(arg[2], deltas)
	default: 
		printHelp()
	}
}

func printHelp() {
	menu := `
*******             **  ** **                    **      **                   **     
/**////**           /** /**//            *****   /**     /**                  /**     
/**   /**   ******  /** /** ** *******  **///**  /**     /**  ******    ******/**     
/*******   **////** /** /**/**//**///**/**  /**  /********** //////**  **//// /****** 
/**///**  /**   /** /** /**/** /**  /**//******  /**//////**  ******* //***** /**///**
/**  //** /**   /** /** /**/** /**  /** /////**  /**     /** **////**  /////**/**  /**
/**   //**//******  *** ***/** ***  /**  *****   /**     /**//******** ****** /**  /**
//     //  //////  /// /// // ///   //  /////    //      //  //////// //////  //   // 
			---- Asaduzzaman Pavel ----

Arguments: 
  - signature old-file signature-file
  - delta signature-file new-file delta-file
`
	fmt.Println(menu)
}
