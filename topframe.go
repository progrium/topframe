package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

	_ "embed"

	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
	"github.com/progrium/macdriver/webkit"
	"github.com/progrium/watcher"
)

var (
	flagEdit  bool
	flagAgent bool

	dir string
)

func init() {
	runtime.LockOSThread()
	flag.BoolVar(&flagEdit, "edit", false, "edit your topframe")
	flag.BoolVar(&flagAgent, "agent", false, "generate agent plist")
}

func edit() {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		log.Println("unable to edit, no value for EDITOR")
		return
	}
	if err := exec.Command(editor, dir).Run(); err != nil {
		log.Println(err)
	}
}

func main() {
	flag.Parse()

	var err error
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dir = filepath.Join(usr.HomeDir, ".topframe")
	os.MkdirAll(dir, 0755)

	//go:embed index.html
	var defaultIndex []byte
	if _, err := os.Stat(filepath.Join(dir, "index.html")); os.IsNotExist(err) {
		ioutil.WriteFile(filepath.Join(dir, "index.html"), defaultIndex, 0644)
	}

	if flagAgent {
		//go:embed agent.plist
		var plist string

		tmpl, err := template.New("plist").Parse(plist)
		if err != nil {
			log.Fatal(err)
		}
		bin, _ := filepath.Abs(os.Args[0])
		err = tmpl.Execute(os.Stdout, struct {
			Dir, Bin string
		}{
			Dir: dir,
			Bin: bin,
		})
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if flagEdit {
		go edit()
	}

	srv := http.Server{
		Handler: http.FileServer(http.Dir(dir)),
	}

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	go srv.Serve(ln)

	fw := watcher.New()
	if err := fw.AddRecursive(dir); err != nil {
		log.Fatal(err)
	}

	go fw.Start(400 * time.Millisecond)

	cocoa.TerminateAfterWindowsClose = false
	app := cocoa.NSApp_WithDidLaunch(func(notification objc.Object) {
		config := webkit.WKWebViewConfiguration_New()
		config.Preferences().SetValueForKey(core.True, core.String("developerExtrasEnabled"))

		wv := webkit.WKWebView_Init(cocoa.NSScreen_Main().Frame(), config)
		wv.SetOpaque(false)
		wv.SetBackgroundColor(cocoa.NSColor_Clear())
		wv.SetValueForKey(core.False, core.String("drawsBackground"))

		url := core.URL(fmt.Sprintf("http://localhost:%d", ln.Addr().(*net.TCPAddr).Port))
		req := core.NSURLRequest_Init(url)
		wv.LoadRequest(req)

		w := cocoa.NSWindow_Init(cocoa.NSScreen_Main().Frame(), cocoa.NSClosableWindowMask,
			cocoa.NSBackingStoreBuffered, false)
		w.SetContentView(wv)
		w.SetBackgroundColor(cocoa.NSColor_Clear())
		w.SetOpaque(false)
		w.SetTitleVisibility(cocoa.NSWindowTitleHidden)
		w.SetTitlebarAppearsTransparent(true)
		w.SetIgnoresMouseEvents(true)
		w.SetLevel(cocoa.NSMainMenuWindowLevel + 2)
		w.MakeKeyAndOrderFront(w)
		w.SetCollectionBehavior(cocoa.NSWindowCollectionBehaviorCanJoinAllSpaces)

		statusBar := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
		statusBar.Retain()
		statusBar.Button().SetTitle("🔲")

		itemInteract := cocoa.NSMenuItem_New()
		itemInteract.Retain()
		itemInteract.SetTitle("Interactive")
		itemInteract.SetAction(objc.Sel("interact:"))
		cocoa.DefaultDelegateClass.AddMethod("interact:", func(_ objc.Object) {
			if w.IgnoresMouseEvents() {
				fmt.Println("Mouse events on")
				w.SetLevel(cocoa.NSMainMenuWindowLevel - 1)
				w.SetIgnoresMouseEvents(false)
				itemInteract.SetState(1)
			} else {
				fmt.Println("Mouse events off")
				w.SetIgnoresMouseEvents(true)
				w.SetLevel(cocoa.NSMainMenuWindowLevel + 2)
				itemInteract.SetState(0)
			}
		})

		itemEnabled := cocoa.NSMenuItem_New()
		itemEnabled.Retain()
		itemEnabled.SetTitle("Enabled")
		itemEnabled.SetState(1)
		itemEnabled.SetAction(objc.Sel("enabled:"))
		cocoa.DefaultDelegateClass.AddMethod("enabled:", func(_ objc.Object) {
			if w.IsVisible() {
				w.Send("orderOut:", w)
				itemInteract.SetEnabled(false)
				itemEnabled.SetState(0)
			} else {
				w.Send("orderFront:", w)
				itemInteract.SetEnabled(true)
				itemEnabled.SetState(1)
			}
		})

		itemQuit := cocoa.NSMenuItem_New()
		itemQuit.SetTitle("Quit")
		itemQuit.SetAction(objc.Sel("terminate:"))

		menu := cocoa.NSMenu_New()
		menu.SetAutoenablesItems(false)
		menu.AddItem(itemEnabled)
		menu.AddItem(itemInteract)

		if os.Getenv("EDITOR") != "" {
			itemEdit := cocoa.NSMenuItem_New()
			itemEdit.SetTitle("Edit Source...")
			itemEdit.SetAction(objc.Sel("edit:"))
			cocoa.DefaultDelegateClass.AddMethod("edit:", func(_ objc.Object) {
				edit()
			})
			menu.AddItem(itemEdit)
		}

		menu.AddItem(cocoa.NSMenuItem_Separator())
		menu.AddItem(itemQuit)
		statusBar.SetMenu(menu)

		go func() {
			for {
				select {
				case event := <-fw.Event:
					if event.IsDir() {
						continue
					}
					wv.Reload(nil)
				case <-fw.Closed:
					return
				}
			}
		}()
	})
	app.ActivateIgnoringOtherApps(true)

	log.Printf("topframe 0.2.0 by progrium\n")
	app.Run()
}
