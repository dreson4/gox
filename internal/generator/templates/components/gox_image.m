// gox_image.m — GOX Image component (UIImageView) with caching + events

#import <UIKit/UIKit.h>
#import <objc/runtime.h>

// Forward declarations from bridge_core.m
@interface GoxRenderContext : NSObject
@property (nonatomic, strong) NSMutableArray *eventHandlers;
@end

typedef struct {
    const char *tag;
    UIView* (*createView)(NSDictionary *props);
    void (*applyStyle)(UIView *view, NSDictionary *props);
    void (*setContent)(UIView *view, NSString *text, NSDictionary *props);
    void (*wireEvent)(UIView *view, int viewID, GoxRenderContext *ctx);
    void (*updateView)(UIView *view, NSDictionary *oldProps, NSDictionary *newProps);
    void (*childAdded)(UIView *parent, UIView *child, CGRect childFrame);
} GoxComponentDef;

extern UIColor* goxParseColor(NSString *hex);
extern void goxRegisterComponent(GoxComponentDef def);
extern void goxLoadImageAsync(UIImageView *imageView, NSString *src, void (^completion)(BOOL success));

// Go exports
extern void GoxHandleLoad(int viewID);
extern void GoxHandleError(int viewID);

// Associated object keys
static const char *kGoxViewID = "gox_viewID";
static const char *kGoxSrc = "gox_src";
static const char *kGoxHasOnLoad = "gox_hasOnLoad";
static const char *kGoxHasOnError = "gox_hasOnError";
static const char *kGoxSpinner = "gox_spinner";

// --- Helpers ---

static void applyContentMode(UIImageView *imageView, NSString *mode) {
    if ([mode isEqualToString:@"cover"]) imageView.contentMode = UIViewContentModeScaleAspectFill;
    else if ([mode isEqualToString:@"stretch"]) imageView.contentMode = UIViewContentModeScaleToFill;
    else if ([mode isEqualToString:@"center"]) imageView.contentMode = UIViewContentModeCenter;
    else imageView.contentMode = UIViewContentModeScaleAspectFit; // "contain" (default)
}

static void showSpinner(UIImageView *imageView) {
    UIActivityIndicatorView *spinner = objc_getAssociatedObject(imageView, kGoxSpinner);
    if (!spinner) {
        spinner = [[UIActivityIndicatorView alloc] initWithActivityIndicatorStyle:UIActivityIndicatorViewStyleMedium];
        spinner.hidesWhenStopped = YES;
        spinner.autoresizingMask = UIViewAutoresizingFlexibleLeftMargin | UIViewAutoresizingFlexibleRightMargin
            | UIViewAutoresizingFlexibleTopMargin | UIViewAutoresizingFlexibleBottomMargin;
        [imageView addSubview:spinner];
        objc_setAssociatedObject(imageView, kGoxSpinner, spinner, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
    }
    spinner.center = CGPointMake(imageView.bounds.size.width / 2, imageView.bounds.size.height / 2);
    [spinner startAnimating];
}

static void hideSpinner(UIImageView *imageView) {
    UIActivityIndicatorView *spinner = objc_getAssociatedObject(imageView, kGoxSpinner);
    if (spinner) [spinner stopAnimating];
}

static void loadImage(UIImageView *imageView, NSString *src, NSDictionary *props) {
    // Store current src for stale-request detection
    objc_setAssociatedObject(imageView, kGoxSrc, src, OBJC_ASSOCIATION_COPY_NONATOMIC);

    // Show placeholder if provided
    NSString *placeholder = props[@"placeholder"];
    if (placeholder && [placeholder length] > 0) {
        UIImage *phImage = [UIImage imageNamed:placeholder];
        if (phImage) imageView.image = phImage;
    }

    // Show activity indicator for remote URLs
    BOOL isRemote = [src hasPrefix:@"http://"] || [src hasPrefix:@"https://"];
    NSNumber *showIndicator = props[@"showActivityIndicator"];
    if (isRemote && (!showIndicator || [showIndicator boolValue])) {
        showSpinner(imageView);
    }

    // Get fade duration (default 0.3 for remote, 0 for local)
    NSNumber *fadeProp = props[@"fadeDuration"];
    CGFloat fadeDuration = isRemote ? 0.3 : 0.0;
    if (fadeProp) fadeDuration = [fadeProp doubleValue];

    NSString *srcCopy = [src copy];
    __unsafe_unretained UIImageView *weakView = imageView;

    goxLoadImageAsync(imageView, src, ^(BOOL success) {
        if (!weakView) return;

        // Stale request check — if src changed since we started loading, ignore
        NSString *currentSrc = objc_getAssociatedObject(weakView, kGoxSrc);
        if (![srcCopy isEqualToString:currentSrc]) return;

        hideSpinner(weakView);

        // Fade-in animation
        if (success && fadeDuration > 0) {
            weakView.alpha = 0;
            [UIView animateWithDuration:fadeDuration animations:^{
                weakView.alpha = 1;
            }];
        }

        // Fire onLoad/onError callbacks
        NSNumber *viewIDNum = objc_getAssociatedObject(weakView, kGoxViewID);
        if (viewIDNum) {
            int viewID = [viewIDNum intValue];
            BOOL hasOnLoad = [objc_getAssociatedObject(weakView, kGoxHasOnLoad) boolValue];
            BOOL hasOnError = [objc_getAssociatedObject(weakView, kGoxHasOnError) boolValue];
            if (success && hasOnLoad) {
                GoxHandleLoad(viewID);
            } else if (!success && hasOnError) {
                GoxHandleError(viewID);
            }
        }
    });
}

// --- Component callbacks ---

static UIView* imageCreate(NSDictionary *props) {
    UIImageView *imageView = [[UIImageView alloc] init];
    imageView.clipsToBounds = YES;
    imageView.userInteractionEnabled = YES;

    applyContentMode(imageView, props[@"contentMode"]);

    NSString *tintColor = props[@"tintColor"];
    if (tintColor) {
        UIColor *c = goxParseColor(tintColor);
        if (c) {
            imageView.tintColor = c;
            imageView.image = [imageView.image imageWithRenderingMode:UIImageRenderingModeAlwaysTemplate];
        }
    }

    NSString *src = props[@"src"];
    if (src && [src length] > 0) {
        loadImage(imageView, src, props);
    }

    return imageView;
}

static void imageApplyStyle(UIView *view, NSDictionary *props) {
    if (![view isKindOfClass:[UIImageView class]]) return;
    UIImageView *imageView = (UIImageView *)view;

    // Re-apply contentMode in case it's in the style props
    NSString *contentMode = props[@"contentMode"];
    if (contentMode) applyContentMode(imageView, contentMode);

    NSString *tintColor = props[@"tintColor"];
    if (tintColor) {
        UIColor *c = goxParseColor(tintColor);
        if (c) {
            imageView.tintColor = c;
            if (imageView.image) {
                imageView.image = [imageView.image imageWithRenderingMode:UIImageRenderingModeAlwaysTemplate];
            }
        }
    }
}

static void imageWireEvent(UIView *view, int viewID, GoxRenderContext *ctx) {
    if (![view isKindOfClass:[UIImageView class]]) return;

    // Store viewID and event flags for the async completion block
    objc_setAssociatedObject(view, kGoxViewID, @(viewID), OBJC_ASSOCIATION_RETAIN_NONATOMIC);

    NSDictionary *props = objc_getAssociatedObject(view, "gox_props");
    objc_setAssociatedObject(view, kGoxHasOnLoad,
        props[@"_hasOnLoad"] ?: @NO, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
    objc_setAssociatedObject(view, kGoxHasOnError,
        props[@"_hasOnError"] ?: @NO, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
}

static void imageUpdate(UIView *view, NSDictionary *oldProps, NSDictionary *newProps) {
    if (![view isKindOfClass:[UIImageView class]]) return;
    UIImageView *imageView = (UIImageView *)view;

    // Store props for wireEvent
    objc_setAssociatedObject(view, "gox_props", newProps, OBJC_ASSOCIATION_RETAIN_NONATOMIC);

    // Update content mode
    NSString *newMode = newProps[@"contentMode"];
    NSString *oldMode = oldProps[@"contentMode"];
    if (newMode && ![newMode isEqual:oldMode]) {
        applyContentMode(imageView, newMode);
    }

    // Update tint color
    NSString *newTint = newProps[@"tintColor"];
    NSString *oldTint = oldProps[@"tintColor"];
    if (newTint && ![newTint isEqual:oldTint]) {
        UIColor *c = goxParseColor(newTint);
        if (c) {
            imageView.tintColor = c;
            if (imageView.image) {
                imageView.image = [imageView.image imageWithRenderingMode:UIImageRenderingModeAlwaysTemplate];
            }
        }
    }

    // Only reload if src changed
    NSString *newSrc = newProps[@"src"];
    NSString *oldSrc = oldProps[@"src"];
    if (newSrc && ![newSrc isEqual:oldSrc]) {
        loadImage(imageView, newSrc, newProps);
    }
}

__attribute__((constructor))
static void goxRegisterImage(void) {
    goxRegisterComponent((GoxComponentDef){
        .tag = "Image",
        .createView = imageCreate,
        .applyStyle = imageApplyStyle,
        .wireEvent = imageWireEvent,
        .updateView = imageUpdate,
    });
}
