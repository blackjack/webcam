go-webcam
=========

Golang webcam wrapper. It depends on v4l2 framework (works only on Linux)

    import "github.com/blackjack/webcam"
    import "fmt"

    func main() {
        fmt.Println(webcam.GetImg("/dev/video0"))
    }
