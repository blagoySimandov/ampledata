import {
  NavbarMain,
  NavbarContent,
  NavbarBrand,
  NavbarLogo,
  NavbarActions,
} from "./navbar";
import { Container } from "@/components/layout/container";
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
      <Container>
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
                  <NavigationMenuLink href="#docs" className="text-base">
                    Docs
                  </NavigationMenuLink>
                </NavigationMenuItem>
              </NavigationMenuList>
            </NavigationMenu>
          </NavbarBrand>

          <NavbarActions>
            <Button variant="ghost">
              <a href="http://localhost:8080/login" className="text-base">
                Login
              </a>
            </Button>
          </NavbarActions>
        </NavbarContent>
      </Container>
    </NavbarMain>
  );
}
