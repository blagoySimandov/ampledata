import {
  NavbarMain,
  NavbarContainer,
  NavbarContent,
  NavbarBrand,
  NavbarLogo,
  NavbarActions,
} from "./navbar";
import { Button } from "@/components/ui/button";
import {
  NavigationMenu,
  NavigationMenuList,
  NavigationMenuItem,
  NavigationMenuLink,
} from "@/components/ui/navigation-menu";
import logo from "@/assets/ampledata-high-resolution-logo-transparent.png";

export function Navbar() {
  return (
    <NavbarMain>
      <NavbarContainer>
        <NavbarContent>
          <NavbarBrand>
            <NavbarLogo>
              <img src={logo} alt="AmpleData" className="h-8" />
            </NavbarLogo>
            <NavigationMenu className="hidden md:flex">
              <NavigationMenuList>
                <NavigationMenuItem>
                  <NavigationMenuLink href="#features" className="text-base">
                    Features
                  </NavigationMenuLink>
                </NavigationMenuItem>
                <NavigationMenuItem>
                  <NavigationMenuLink href="#pricing" className="text-base">
                    Pricing
                  </NavigationMenuLink>
                </NavigationMenuItem>
                <NavigationMenuItem>
                  <NavigationMenuLink href="#docs" className="text-base">Docs</NavigationMenuLink>
                </NavigationMenuItem>
              </NavigationMenuList>
            </NavigationMenu>
          </NavbarBrand>

          <NavbarActions>
            <Button variant="ghost" asChild>
              <a href="#login" className="text-base">Login</a>
            </Button>
            <Button asChild>
              <a href="#signup" className="text-base">Sign Up</a>
            </Button>
          </NavbarActions>
        </NavbarContent>
      </NavbarContainer>
    </NavbarMain>
  );
}
