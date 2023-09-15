const {
    AppDelegate,
    UIApplication,
    NSURL,
    NSDictionary,
    UIWindowScene,
    NSUserActivity,
} = ObjC.classes;

var app = null;
var delegate = null;
var shared = null;
var scene = null;
var sceneDelegate = null;
var opts = NSDictionary.alloc().init();

var NSUserActivityTypeBrowsingWeb = null;
var activity = null;

var NSUserActivityTypeBrowsingWeb = null
var activity = null;

rpc.exports = {
    setup(method, appName, delegateName) {
        switch (method) {
            case "delegate":
                if (delegate == null) {
                    if (delegateName != "" ) {
                        delegate = ObjC.Object(ObjC.chooseSync(ObjC.classes[delegateName])[0]);
                    } else {
                        delegate = ObjC.Object(ObjC.chooseSync(AppDelegate)[0]);
                    }
                    if (appName != "") {
                        app = ObjC.Object(ObjC.chooseSync(UIApplication)[0]);
                    } else {
                        app = ObjC.Object(ObjC.chooseSync(ObjC.classes[appName])[0]);
                    }
                }
                break;
            case "app":
                if (appName != "") {
                    app = ObjC.Object(ObjC.chooseSync(UIApplication)[0]);
                } else {
                    app = ObjC.Object(ObjC.chooseSync(ObjC.classes[appName])[0]);
                }
                break;
            case "scene_activity":
                if (shared == null) {
                    NSUserActivityTypeBrowsingWeb = ObjC.Object(Memory.readPointer(Module.findExportByName(null, "NSUserActivityTypeBrowsingWeb")));
                    activity = NSUserActivity.alloc().initWithActivityType_(NSUserActivityTypeBrowsingWeb);
                    sceneDelegate = ObjC.Object(ObjC.chooseSync(ObjC.classes[delegateName])[0]);
                    shared = ObjC.Object(UIApplication.sharedApplication());
                    scene = ObjC.Object(ObjC.chooseSync(UIWindowScene)[0]);
                }
            default:
                return "method not implemented";
        }
    },
    fuzz(method, url, appName, delegateName) {
        var ur = NSURL.URLWithString_(url);
        switch (method) {
            case "delegate":
                delegate.application_openURL_options_(app, ur, opts);
                break;
            case "app":
                app.openURL_(ur);
                break;
            case "scene_activity":
                activity.setWebPageURL_(ur);

                ObjC.schedule(ObjC.mainQueue, () => {
                    sceneDelegate.scene_continueUserActivity_(scene, activity);
                })
            default:
                return "method not implemented";
        }
    }
};