import { cva } from "../../styled-system/css";

export const buttonRecipe = cva({
  base: {
    display: "inline-flex",
    alignItems: "center",
    justifyContent: "center",
    gap: "2",
    fontWeight: "semibold",
    letterSpacing: "tight",
    outline: "none",
    transitionProperty: "common",
    transitionDuration: "normal",
    _focusVisible: { outlineWidth: "2px", outlineColor: "brand.300" },
    _active: { transform: "scale(0.985)" },
    _disabled: {
      pointerEvents: "none",
      opacity: 0.5,
    },
  },
  variants: {
    variant: {
      primary: {
        bg: "brand.600",
        color: "white",
        boxShadow: "sm",
        _hover: { bg: "brand.700" },
      },
      secondary: {
        borderWidth: "1px",
        borderColor: "gray.200",
        bg: "gray.100",
        color: "gray.700",
        boxShadow: "sm",
        _hover: { bg: "gray.200" },
      },
      ghost: {
        color: "gray.600",
        _hover: { bg: "gray.100", color: "gray.900" },
      },
      danger: {
        borderWidth: "1px",
        borderColor: "red.200",
        bg: "red.50",
        color: "red.700",
        boxShadow: "sm",
        _hover: { bg: "red.100" },
      },
      text: {
        color: "gray.600",
        _hover: { color: "gray.900" },
      },
    },
    size: {
      sm: { minH: "9", borderRadius: "lg", px: "3", py: "2", fontSize: "xs" },
      md: { minH: "10", borderRadius: "xl", px: "4", py: "2.5", fontSize: "sm" },
      lg: { minH: "12", borderRadius: "2xl", px: "6", py: "3", fontSize: "sm" },
      icon: { h: "10", w: "10", borderRadius: "xl", p: "0" },
    },
    block: {
      true: { w: "full" },
    },
  },
  defaultVariants: {
    variant: "primary",
    size: "md",
  },
});

export const fieldRecipe = cva({
  base: {
    w: "full",
    borderRadius: "xl",
    borderWidth: "1px",
    borderColor: "gray.200",
    bg: "gray.50",
    color: "gray.900",
    boxShadow: "sm",
    outline: "none",
    transitionProperty: "common",
    transitionDuration: "normal",
    _placeholder: { color: "gray.400" },
    _focus: { borderColor: "brand.300", outlineWidth: "2px", outlineColor: "brand.200" },
    _disabled: {
      cursor: "not-allowed",
      opacity: 0.6,
    },
  },
  variants: {
    size: {
      sm: { minH: "10", px: "3", py: "2", fontSize: "sm" },
      md: { minH: "11", px: "4", py: "2.5", fontSize: "sm" },
      area: { px: "4", py: "3", fontSize: "sm" },
    },
    mono: {
      true: { fontFamily: "mono" },
    },
  },
  defaultVariants: {
    size: "md",
  },
});
