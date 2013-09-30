go-webcam
=========

Golang webcam wrapper

    import "github.com/blackjack/webcam"
    import "fmt"

    func main() {
        fmt.Println(webcam.GetImg("/dev/video0"))
    }
