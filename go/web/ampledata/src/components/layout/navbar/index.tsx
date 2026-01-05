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
import { useAuth } from "@workos-inc/authkit-react";
import logo from "@/assets/ampledata-high-resolution-logo-transparent.png";

export function Navbar() {
  const { user, isLoading, signIn, signUp, signOut } = useAuth();

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
            {isLoading ? (
              <p className="text-sm text-muted-foreground">Loading...</p>
            ) : user ? (
              <>
                <p className="text-sm hidden sm:block">
                  {user.firstName || user.email}
                </p>
                <Button
                  variant="ghost"
                  onClick={() => {
                    signOut();
                  }}
                >
                  Sign Out
                </Button>
              </>
            ) : (
              <>
                <Button
                  variant="ghost"
                  onClick={() => {
                    signIn();
                  }}
                >
                  Login
                </Button>
                <Button
                  onClick={() => {
                    signUp();
                  }}
                >
                  Sign Up
                </Button>
              </>
            )}
          </NavbarActions>
        </NavbarContent>
      </Container>
    </NavbarMain>
  );
}
