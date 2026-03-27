// gox_platform.m — Platform API implementations for iOS
// Provides native implementations for storage, clipboard, haptics,
// linking, keyboard, and status bar APIs.

#import <UIKit/UIKit.h>

// ============================================================
// Storage — NSUserDefaults
// ============================================================

int goxStorageSet(const char *key, const char *value) {
    @autoreleasepool {
        NSString *k = [NSString stringWithUTF8String:key];
        NSString *v = [NSString stringWithUTF8String:value];
        [[NSUserDefaults standardUserDefaults] setObject:v forKey:k];
        return 0;
    }
}

// Returns a C string that must be freed by the caller via GoxFreeString.
const char* goxStorageGet(const char *key) {
    @autoreleasepool {
        NSString *k = [NSString stringWithUTF8String:key];
        NSString *v = [[NSUserDefaults standardUserDefaults] stringForKey:k];
        if (v == nil) return NULL;
        return strdup([v UTF8String]);
    }
}

void goxStorageDelete(const char *key) {
    @autoreleasepool {
        NSString *k = [NSString stringWithUTF8String:key];
        [[NSUserDefaults standardUserDefaults] removeObjectForKey:k];
    }
}

// ============================================================
// Clipboard — UIPasteboard
// ============================================================

void goxClipboardCopy(const char *text) {
    @autoreleasepool {
        NSString *t = [NSString stringWithUTF8String:text];
        [UIPasteboard generalPasteboard].string = t;
    }
}

// Returns a C string that must be freed by the caller via GoxFreeString.
const char* goxClipboardRead(void) {
    @autoreleasepool {
        NSString *t = [UIPasteboard generalPasteboard].string;
        if (t == nil) return NULL;
        return strdup([t UTF8String]);
    }
}

// ============================================================
// Haptics — UIFeedbackGenerator
// ============================================================

void goxHapticImpact(int style) {
    @autoreleasepool {
        UIImpactFeedbackStyle s;
        switch (style) {
            case 0: s = UIImpactFeedbackStyleLight; break;
            case 1: s = UIImpactFeedbackStyleMedium; break;
            case 2: s = UIImpactFeedbackStyleHeavy; break;
            default: s = UIImpactFeedbackStyleMedium; break;
        }
        dispatch_async(dispatch_get_main_queue(), ^{
            UIImpactFeedbackGenerator *gen = [[UIImpactFeedbackGenerator alloc] initWithStyle:s];
            [gen prepare];
            [gen impactOccurred];
        });
    }
}

void goxHapticNotify(int typ) {
    @autoreleasepool {
        UINotificationFeedbackType t;
        switch (typ) {
            case 0: t = UINotificationFeedbackTypeSuccess; break;
            case 1: t = UINotificationFeedbackTypeWarning; break;
            case 2: t = UINotificationFeedbackTypeError; break;
            default: t = UINotificationFeedbackTypeSuccess; break;
        }
        dispatch_async(dispatch_get_main_queue(), ^{
            UINotificationFeedbackGenerator *gen = [[UINotificationFeedbackGenerator alloc] init];
            [gen prepare];
            [gen notificationOccurred:t];
        });
    }
}

void goxHapticSelection(void) {
    @autoreleasepool {
        dispatch_async(dispatch_get_main_queue(), ^{
            UISelectionFeedbackGenerator *gen = [[UISelectionFeedbackGenerator alloc] init];
            [gen prepare];
            [gen selectionChanged];
        });
    }
}

// ============================================================
// Linking — UIApplication openURL
// ============================================================

int goxOpenURL(const char *url) {
    @autoreleasepool {
        NSString *u = [NSString stringWithUTF8String:url];
        NSURL *nsURL = [NSURL URLWithString:u];
        if (nsURL == nil) return 1;
        dispatch_async(dispatch_get_main_queue(), ^{
            [[UIApplication sharedApplication] openURL:nsURL options:@{} completionHandler:nil];
        });
        return 0;
    }
}

// ============================================================
// Keyboard — dismiss
// ============================================================

void goxDismissKeyboard(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        UIWindow *window = nil;
        for (UIWindowScene *scene in [UIApplication sharedApplication].connectedScenes) {
            if (scene.activationState == UISceneActivationStateForegroundActive) {
                for (UIWindow *w in scene.windows) {
                    if (w.isKeyWindow) {
                        window = w;
                        break;
                    }
                }
            }
        }
        if (window) {
            [window endEditing:YES];
        }
    });
}

// ============================================================
// Status Bar — style
// ============================================================

static int goxStatusBarStyle = 0; // 0 = auto, 1 = light, 2 = dark

void goxSetStatusBar(const char *style) {
    @autoreleasepool {
        NSString *s = [NSString stringWithUTF8String:style];
        if ([s isEqualToString:@"light"]) {
            goxStatusBarStyle = 1;
        } else if ([s isEqualToString:@"dark"]) {
            goxStatusBarStyle = 2;
        } else {
            goxStatusBarStyle = 0;
        }
        dispatch_async(dispatch_get_main_queue(), ^{
            [UIApplication.sharedApplication.connectedScenes enumerateObjectsUsingBlock:^(UIScene *scene, BOOL *stop) {
                if ([scene isKindOfClass:[UIWindowScene class]]) {
                    UIWindowScene *ws = (UIWindowScene *)scene;
                    for (UIWindow *w in ws.windows) {
                        [w.rootViewController setNeedsStatusBarAppearanceUpdate];
                    }
                }
            }];
        });
    }
}

// Called by the root view controller to query the current style.
int goxGetStatusBarStyle(void) {
    return goxStatusBarStyle;
}
