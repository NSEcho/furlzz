const {
    AppDelegate,
    UIApplication,
    NSURL,
    NSDictionary,
    UIWindowScene,
    NSUserActivity,
} = ObjC.classes;

var delegate = null;
var shared = null;
var scene = null;
var sceneDelegate = null;
var opts = NSDictionary.alloc().init();
const app = ObjC.Object(ObjC.chooseSync(UIApplication)[0]);

const NSUserActivityTypeBrowsingWeb = ObjC.Object(Memory.readPointer(Module.findExportByName(null, "NSUserActivityTypeBrowsingWeb")));
var activity = NSUserActivity.alloc().initWithActivityType_(NSUserActivityTypeBrowsingWeb);

rpc.exports = {
    fuzz(method, url, delegateName) {
        var ur = NSURL.URLWithString_(url);
        switch (method) {
            case "delegate":
                if (delegate == null) {
                    delegate = ObjC.Object(ObjC.chooseSync(AppDelegate)[0]);
                }
                delegate.application_openURL_options_(app, ur, opts);
                break;
            case "app":
                app.openURL_(ur);
            case "scene_activity":
                if (shared == null) {
                    sceneDelegate = ObjC.Object(ObjC.chooseSync(ObjC.classes[delegateName])[0]);
                    shared = ObjC.Object(UIApplication.sharedApplication());
                    scene = ObjC.Object(ObjC.chooseSync(UIWindowScene)[0]);
                }
                activity.setWebPageURL_(ur);

                ObjC.schedule(ObjC.mainQueue, () => {
                    sceneDelegate.scene_continueUserActivity_(scene, activity);
                })
            default:
                return "method not implemented";
        }
    }
};