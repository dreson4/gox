// gox_image.m — GOX Image component (UIImageView)

#import <UIKit/UIKit.h>

@class GoxRenderContext;

typedef struct {
    const char *tag;
    UIView* (*createView)(NSDictionary *props);
    void (*applyStyle)(UIView *view, NSDictionary *props);
    void (*setContent)(UIView *view, NSString *text, NSDictionary *props);
    void (*wireEvent)(UIView *view, int viewID, GoxRenderContext *ctx);
    void (*updateView)(UIView *view, NSDictionary *oldProps, NSDictionary *newProps);
    void (*childAdded)(UIView *parent, UIView *child, CGRect childFrame);
} GoxComponentDef;

extern void goxLoadImageAsync(UIImageView *imageView, NSString *src);
extern void goxRegisterComponent(GoxComponentDef def);

static UIView* imageCreate(NSDictionary *props) {
    UIImageView *imageView = [[UIImageView alloc] init];
    imageView.contentMode = UIViewContentModeScaleAspectFit;
    imageView.clipsToBounds = YES;

    NSString *contentMode = props[@"contentMode"];
    if ([contentMode isEqualToString:@"cover"]) imageView.contentMode = UIViewContentModeScaleAspectFill;
    else if ([contentMode isEqualToString:@"stretch"]) imageView.contentMode = UIViewContentModeScaleToFill;
    else if ([contentMode isEqualToString:@"center"]) imageView.contentMode = UIViewContentModeCenter;

    goxLoadImageAsync(imageView, props[@"src"]);
    return imageView;
}

static void imageUpdate(UIView *view, NSDictionary *oldProps, NSDictionary *newProps) {
    if (![view isKindOfClass:[UIImageView class]]) return;

    NSString *newSrc = newProps[@"src"];
    NSString *oldSrc = oldProps[@"src"];
    if (![newSrc isEqual:oldSrc]) {
        goxLoadImageAsync((UIImageView *)view, newSrc);
    }
}

__attribute__((constructor))
static void goxRegisterImage(void) {
    goxRegisterComponent((GoxComponentDef){
        .tag = "Image",
        .createView = imageCreate,
        .updateView = imageUpdate,
    });
}
