// gox_image.m — GOX Image component powered by SDWebImage

#import <UIKit/UIKit.h>
#import <objc/runtime.h>
#import "SDWebImage.h"

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

// Go exports
extern void GoxHandleLoad(int viewID);
extern void GoxHandleError(int viewID);

// Associated object keys
static const char *kGoxViewID = "gox_viewID";
static const char *kGoxHasOnLoad = "gox_hasOnLoad";
static const char *kGoxHasOnError = "gox_hasOnError";

// --- Helpers ---

static void applyContentMode(UIImageView *imageView, NSString *mode) {
    if ([mode isEqualToString:@"cover"]) imageView.contentMode = UIViewContentModeScaleAspectFill;
    else if ([mode isEqualToString:@"stretch"]) imageView.contentMode = UIViewContentModeScaleToFill;
    else if ([mode isEqualToString:@"center"]) imageView.contentMode = UIViewContentModeCenter;
    else imageView.contentMode = UIViewContentModeScaleAspectFit; // "contain" (default)
}

static void loadImage(UIImageView *imageView, NSString *src, NSDictionary *props) {
    if (!src || [src length] == 0) return;

    // Local asset — load directly
    if (![src hasPrefix:@"http://"] && ![src hasPrefix:@"https://"]) {
        imageView.image = [UIImage imageNamed:src];
        return;
    }

    // Placeholder image
    UIImage *placeholderImage = nil;
    NSString *placeholder = props[@"placeholder"];
    if (placeholder && [placeholder length] > 0) {
        placeholderImage = [UIImage imageNamed:placeholder];
    }

    // Build SDWebImage options
    SDWebImageOptions options = 0;

    // Fade transition
    NSNumber *fadeProp = props[@"fadeDuration"];
    CGFloat fadeDuration = 0.3;
    if (fadeProp) fadeDuration = [fadeProp doubleValue];

    NSURL *url = [NSURL URLWithString:src];

    [imageView sd_setImageWithURL:url
                 placeholderImage:placeholderImage
                          options:options
                        completed:^(UIImage *image, NSError *error, SDImageCacheType cacheType, NSURL *imageURL) {
        // Fire onLoad/onError callbacks
        NSNumber *viewIDNum = objc_getAssociatedObject(imageView, kGoxViewID);
        if (!viewIDNum) return;

        int viewID = [viewIDNum intValue];

        if (image && !error) {
            BOOL hasOnLoad = [objc_getAssociatedObject(imageView, kGoxHasOnLoad) boolValue];
            if (hasOnLoad) GoxHandleLoad(viewID);
        } else {
            BOOL hasOnError = [objc_getAssociatedObject(imageView, kGoxHasOnError) boolValue];
            if (hasOnError) GoxHandleError(viewID);
        }
    }];

    // Apply fade transition for non-cached images
    if (fadeDuration > 0) {
        imageView.sd_imageTransition = [SDWebImageTransition fadeTransitionWithDuration:fadeDuration];
    }

    // Activity indicator
    NSNumber *showIndicator = props[@"showActivityIndicator"];
    if (!showIndicator || [showIndicator boolValue]) {
        imageView.sd_imageIndicator = SDWebImageActivityIndicator.grayIndicator;
    } else {
        imageView.sd_imageIndicator = nil;
    }
}

// --- Component callbacks ---

static UIView* imageCreate(NSDictionary *props) {
    UIImageView *imageView = [[UIImageView alloc] init];
    imageView.clipsToBounds = YES;
    imageView.userInteractionEnabled = YES;

    applyContentMode(imageView, props[@"contentMode"]);

    // Store props for wireEvent
    objc_setAssociatedObject(imageView, "gox_props", props, OBJC_ASSOCIATION_RETAIN_NONATOMIC);

    NSString *src = props[@"src"];
    if (src && [src length] > 0) {
        loadImage(imageView, src, props);
    }

    return imageView;
}

static void imageApplyStyle(UIView *view, NSDictionary *props) {
    if (![view isKindOfClass:[UIImageView class]]) return;
    UIImageView *imageView = (UIImageView *)view;

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

    // Only reload if src changed — SDWebImage handles caching
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
