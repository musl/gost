
package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
)

/*
 * The capacity of this channel must be equal to the number of important
 * groutines that this program starts.  In each go routine, the first 
 * task is to push an integer onto the channel, and the last task is to
 * pop one off.  If this convention is maintained, the length of the
 * channel is the number of goroutines in service, and if the channel's
 * length and capacity are equal, this service is healthy.
 */
var service_status = make(chan int, 2)

/*
 * Configure anything that needs configuring.
 */
func receive_configuration() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

/*
 * Convenience method to log a particular request.
 */
func log_request(req *http.Request) {
	log.Printf("%s %s from %s ", req.Method, req.RequestURI, req.RemoteAddr)
}

/*
 * Spawn off goroutines to handle incoming requests.
 */
func go_serve() {

	/*
	 * App routes.
	 */
	http.HandleFunc("/down", route_down)
	http.HandleFunc("/up", route_up)

	// Status endpoint.
	http.HandleFunc("/status/", route_status)

	// Default, all-maching route.
	http.HandleFunc("/", route_default)

	go func() {
		service_status<- 1
		log.Println("Listening on :8000")
		err := http.ListenAndServe(":8000", nil)
		<-service_status
		log.Fatal(err)
	}()

	go func() {
		service_status<- 1
		log.Println("Listening on :8443")
		err := http.ListenAndServeTLS(":8443", "gost.crt", "gost.key",  nil)
		<-service_status
		log.Fatal(err)
	}()

}

/*
 * Wait for an interrupt signal.
 */
func wait_for_death() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	log.Println("Killed.")
	os.Exit(0)
}

/*
 * A default all-matching route to allow the app's response behavior to
 * be fully defined.
 */
func route_default(res http.ResponseWriter, req *http.Request) {
	log_request(req)

	if(req.URL.Path != "/") {
		res.WriteHeader(404)
		io.WriteString(res, "Not Found")
		return
	}

	io.WriteString(res, "")
}

/*
 * The obligatory status endpoint that's used to determine service
 * health externally.
 */
func route_status(res http.ResponseWriter, req *http.Request) {
	log_request(req)

	if(len(service_status) != cap(service_status)) {
		res.WriteHeader(404)
		io.WriteString(res, "Unhealthy")
		return
	}

	io.WriteString(res, "Healthy")
}

/*
 * GET: Perform a downstream bandwidth test.
 */
func route_down(res http.ResponseWriter, req *http.Request) {
	log_request(req)

	if(req.Method != "GET" && req.Method != "") {
		res.WriteHeader(405) // Method Not Allowed
		io.WriteString(res, "Method Not Allowed")
		return
	}

	io.WriteString(res, "Download Test")
}

/*
 * PUT: Perform an upstream bandwidth test.
 */
func route_up(res http.ResponseWriter, req *http.Request) {
	log_request(req)

	if(req.Method != "PUT" && req.Method != "") {
		res.WriteHeader(405) // Method Not Allowed
		io.WriteString(res, "Method Not Allowed")
		return
	}

	io.WriteString(res, "Upload Test")
}

/*
 * Main entry point and short synopsis of execution flow. 
 */
func main() {
	receive_configuration()
	go_serve()
	wait_for_death()
}

