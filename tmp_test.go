package main
import (
    "bytes"
    "github.com/jung-kurt/gofpdf"
    "os"
)
func main(){
    pdf:=gofpdf.New("P","mm","A4","")
    pdf.AddPage()
    pdf.SetFont("Helvetica","",12)
    pdf.Cell(40,10,"Hello PDF")
    var buf bytes.Buffer
    if err:=pdf.Output(&buf); err!=nil{panic(err)}
    os.WriteFile("/var/www/html/vickygo/output.pdf",buf.Bytes(),0644)
}

