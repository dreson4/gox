// gox_text.m — GOX Text component (UILabel)

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

extern UIColor* goxParseColor(NSString *hex);
extern void goxRegisterComponent(GoxComponentDef def);

static UIView* textCreate(NSDictionary *props) {
    UILabel *label = [[UILabel alloc] init];
    label.numberOfLines = 0;
    return label;
}

static void textApplyStyle(UIView *view, NSDictionary *props) {
    if (![view isKindOfClass:[UILabel class]]) return;
    UILabel *label = (UILabel *)view;

    NSDictionary *style = props[@"style"];
    if (!style) return;

    NSNumber *fontSize = style[@"FontSize"];
    NSString *fontWeight = style[@"FontWeight"];
    CGFloat size = (fontSize && [fontSize doubleValue] > 0) ? [fontSize doubleValue] : 17.0;

    if (fontWeight && [fontWeight length] > 0) {
        UIFontWeight weight = UIFontWeightRegular;
        if ([fontWeight isEqualToString:@"bold"] || [fontWeight isEqualToString:@"700"]) weight = UIFontWeightBold;
        else if ([fontWeight isEqualToString:@"600"]) weight = UIFontWeightSemibold;
        else if ([fontWeight isEqualToString:@"500"]) weight = UIFontWeightMedium;
        else if ([fontWeight isEqualToString:@"300"]) weight = UIFontWeightLight;
        label.font = [UIFont systemFontOfSize:size weight:weight];
    } else if (fontSize && [fontSize doubleValue] > 0) {
        label.font = [UIFont systemFontOfSize:size];
    }

    NSString *color = style[@"Color"];
    if (color && [color length] > 0) {
        UIColor *c = goxParseColor(color);
        if (c) label.textColor = c;
    }

    NSString *textAlign = style[@"TextAlign"];
    if ([textAlign isEqualToString:@"center"]) label.textAlignment = NSTextAlignmentCenter;
    else if ([textAlign isEqualToString:@"right"]) label.textAlignment = NSTextAlignmentRight;
}

static void textSetContent(UIView *view, NSString *text, NSDictionary *props) {
    if ([view isKindOfClass:[UILabel class]]) {
        ((UILabel *)view).text = text;
    }
}

__attribute__((constructor))
static void goxRegisterText(void) {
    GoxComponentDef def = {
        .tag = "Text",
        .createView = textCreate,
        .applyStyle = textApplyStyle,
        .setContent = textSetContent,
    };
    goxRegisterComponent(def);

    // _text uses the same implementation
    GoxComponentDef textDef = def;
    textDef.tag = "_text";
    goxRegisterComponent(textDef);
}
