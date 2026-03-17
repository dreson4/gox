// gox_switch.m — GOX Switch component (UISwitch)

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

static UIView* switchCreate(NSDictionary *props) {
    UISwitch *sw = [[UISwitch alloc] init];
    NSNumber *value = props[@"value"];
    if (value) sw.on = [value boolValue];
    NSNumber *disabled = props[@"disabled"];
    if (disabled && [disabled boolValue]) sw.enabled = NO;
    return sw;
}

__attribute__((constructor))
static void goxRegisterSwitch(void) {
    goxRegisterComponent((GoxComponentDef){
        .tag = "Switch",
        .createView = switchCreate,
    });
}
