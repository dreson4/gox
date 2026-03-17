// gox_button.m — GOX Button component (UIButton)

#import <UIKit/UIKit.h>

// Forward declarations from bridge_core.m
@interface GoxRenderContext : NSObject
@property (nonatomic, strong) NSMutableArray *eventHandlers;
@end

@interface GoxEventHandler : NSObject
@property (nonatomic, assign) int viewID;
- (void)handleTap;
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

static UIView* buttonCreate(NSDictionary *props) {
    UIButton *button = [UIButton buttonWithType:UIButtonTypeCustom];
    return button;
}

static void buttonSetContent(UIView *view, NSString *text, NSDictionary *props) {
    if (![view isKindOfClass:[UIButton class]]) return;
    UIButton *button = (UIButton *)view;

    [button setTitle:text forState:UIControlStateNormal];

    // Apply button text styling from child Text element
    NSString *btnColor = props[@"_btnTextColor"];
    if (btnColor && [btnColor length] > 0) {
        UIColor *c = goxParseColor(btnColor);
        if (c) [button setTitleColor:c forState:UIControlStateNormal];
    }

    NSNumber *btnSize = props[@"_btnTextSize"];
    if (btnSize && [btnSize doubleValue] > 0) {
        CGFloat size = [btnSize doubleValue];
        NSString *btnWeight = props[@"_btnTextWeight"];
        UIFontWeight weight = UIFontWeightRegular;
        if ([btnWeight isEqualToString:@"bold"] || [btnWeight isEqualToString:@"700"]) weight = UIFontWeightBold;
        else if ([btnWeight isEqualToString:@"600"]) weight = UIFontWeightSemibold;
        button.titleLabel.font = [UIFont systemFontOfSize:size weight:weight];
    }
}

static void buttonWireEvent(UIView *view, int viewID, GoxRenderContext *ctx) {
    if (![view isKindOfClass:[UIButton class]]) return;

    // Remove old targets to avoid duplicates on re-render
    [(UIButton *)view removeTarget:nil
                            action:NULL
                  forControlEvents:UIControlEventTouchUpInside];

    GoxEventHandler *handler = [[GoxEventHandler alloc] init];
    handler.viewID = viewID;
    [(UIButton *)view addTarget:handler
                         action:@selector(handleTap)
               forControlEvents:UIControlEventTouchUpInside];
    [ctx.eventHandlers addObject:handler];
}

__attribute__((constructor))
static void goxRegisterButton(void) {
    goxRegisterComponent((GoxComponentDef){
        .tag = "Button",
        .createView = buttonCreate,
        .setContent = buttonSetContent,
        .wireEvent = buttonWireEvent,
    });
}
