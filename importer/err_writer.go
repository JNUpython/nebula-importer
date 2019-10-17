package nebula_importer

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
	"time"
)

type ErrorWriter interface {
	SetErrorHandler()
}

type CSVErrWriter struct {
	ErrConf ErrorConfig
	ErrCh   <-chan ErrData
}

func (w *CSVErrWriter) SetErrorHandler() {
	go func() {
		dataFile, err := os.Create(w.ErrConf.ErrorDataPath)
		if err != nil {
			log.Fatal(err)
		}
		defer dataFile.Close()

		dataWriter := csv.NewWriter(dataFile)

		logFile, err := os.Create(w.ErrConf.ErrorLogPath)
		if err != nil {
			log.Fatal(err)
		}
		defer logFile.Close()

		logWriter := bufio.NewWriter(logFile)

		ticker := time.NewTicker(30 * time.Second)

		var numFailed uint64 = 0
		for {
			select {
			case <-ticker.C:
				log.Printf("Failed queries: %d", numFailed)
			case rawErr := <-w.ErrCh:
				// Write failed data
				errData := make([]string, len(rawErr.Data))
				for i := range rawErr.Data {
					errData[i] = rawErr.Data[i].(string)
				}

				dataWriter.Write(errData)

				// Write error message
				logWriter.WriteString(err.Error())
				logWriter.WriteString("\n")

				numFailed++
			}
		}
	}()

	log.Println("Setup CSV error handler")
}