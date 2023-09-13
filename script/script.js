const {
    AppDelegate,
    UIApplication,
    NSURL,
    NSDictionary,
} = ObjC.classes;

var delegate = null;
var opts = NSDictionary.alloc().init();
const app = ObjC.Object(ObjC.chooseSync(UIApplication)[0]);

rpc.exports = {
    fuzz(method, url) {
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
            default:
                return "method not implemented";
        }
    }
};