// gox_view.m — GOX View component (default container)

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

extern void goxRegisterComponent(GoxComponentDef def);

static UIView* viewCreate(NSDictionary *props) {
    UIView *v = [[UIView alloc] init];
    v.clipsToBounds = YES;
    return v;
}

__attribute__((constructor))
static void goxRegisterView(void) {
    goxRegisterComponent((GoxComponentDef){
        .tag = "View",
        .createView = viewCreate,
    });
    // SafeArea and Fragment are structural — layout handled by Go/core bridge.
    // Register them with the same view factory.
    goxRegisterComponent((GoxComponentDef){
        .tag = "SafeArea",
        .createView = viewCreate,
    });
}
